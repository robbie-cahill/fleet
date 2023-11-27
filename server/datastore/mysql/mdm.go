package mysql

import (
	"context"
	"errors"
	"fmt"

	"github.com/fleetdm/fleet/v4/server/contexts/ctxerr"
	"github.com/fleetdm/fleet/v4/server/fleet"
	"github.com/fleetdm/fleet/v4/server/mdm/apple/mobileconfig"
	"github.com/go-kit/kit/log/level"
	"github.com/jmoiron/sqlx"
)

func (ds *Datastore) GetMDMCommandPlatform(ctx context.Context, commandUUID string) (string, error) {
	stmt := `
SELECT CASE
	WHEN EXISTS (SELECT 1 FROM nano_commands WHERE command_uuid = ?) THEN 'darwin'
	WHEN EXISTS (SELECT 1 FROM windows_mdm_commands WHERE command_uuid = ?) THEN 'windows'
	ELSE ''
END AS platform
`

	var p string
	if err := sqlx.GetContext(ctx, ds.reader(ctx), &p, stmt, commandUUID, commandUUID); err != nil {
		return "", err
	}
	if p == "" {
		return "", ctxerr.Wrap(ctx, notFound("MDMCommand").WithName(commandUUID))
	}

	return p, nil
}

func (ds *Datastore) ListMDMCommands(
	ctx context.Context,
	tmFilter fleet.TeamFilter,
	listOpts *fleet.MDMCommandListOptions,
) ([]*fleet.MDMCommand, error) {
	appleStmt := `
SELECT
    nvq.id as host_uuid,
    nvq.command_uuid,
    COALESCE(NULLIF(nvq.status, ''), 'Pending') as status,
    COALESCE(nvq.result_updated_at, nvq.created_at) as updated_at,
    nvq.request_type,
    h.hostname,
    h.team_id
FROM
    nano_view_queue nvq
INNER JOIN
    hosts h
ON
    nvq.id = h.uuid
WHERE
   nvq.active = 1
`

	windowsStmt := `
SELECT
    mwe.host_uuid,
    wmc.command_uuid,
    COALESCE(NULLIF(wmcr.status_code, ''), 'Pending') as status,
    COALESCE(wmc.updated_at, wmc.created_at) as updated_at,
    wmc.target_loc_uri as request_type,
    h.hostname,
    h.team_id
FROM windows_mdm_commands wmc
LEFT JOIN windows_mdm_command_queue wmcq ON wmcq.command_uuid = wmc.command_uuid
LEFT JOIN windows_mdm_command_results wmcr ON wmc.command_uuid = wmcr.command_uuid
INNER JOIN mdm_windows_enrollments mwe ON wmcq.enrollment_id = mwe.id OR wmcr.enrollment_id = mwe.id
INNER JOIN hosts h ON h.uuid = mwe.host_uuid
`

	jointStmt := fmt.Sprintf(
		`SELECT * FROM ((%s) UNION ALL (%s)) as combined_commands WHERE %s`,
		appleStmt, windowsStmt, ds.whereFilterHostsByTeams(tmFilter, "h"),
	)
	jointStmt, params := appendListOptionsWithCursorToSQL(jointStmt, nil, &listOpts.ListOptions)
	var results []*fleet.MDMCommand
	if err := sqlx.SelectContext(ctx, ds.reader(ctx), &results, jointStmt, params...); err != nil {
		return nil, ctxerr.Wrap(ctx, err, "list commands")
	}
	return results, nil
}

func (ds *Datastore) BatchSetMDMProfiles(ctx context.Context, tmID *uint, macProfiles []*fleet.MDMAppleConfigProfile, winProfiles []*fleet.MDMWindowsConfigProfile) error {
	return ds.withTx(ctx, func(tx sqlx.ExtContext) error {
		if err := ds.batchSetMDMWindowsProfilesDB(ctx, tx, tmID, winProfiles); err != nil {
			return ctxerr.Wrap(ctx, err, "batch set windows profiles")
		}

		if err := ds.batchSetMDMAppleProfilesDB(ctx, tx, tmID, macProfiles); err != nil {
			return ctxerr.Wrap(ctx, err, "batch set apple profiles")
		}

		return nil
	})
}

func (ds *Datastore) ListMDMConfigProfiles(ctx context.Context, teamID *uint, opt fleet.ListOptions) ([]*fleet.MDMConfigProfilePayload, *fleet.PaginationMetadata, error) {

	var profs []*fleet.MDMConfigProfilePayload

	const selectStmt = `
SELECT
	profile_uuid,
	team_id,
	name,
	platform,
	identifier,
	checksum,
	created_at,
	updated_at
FROM (
	SELECT
		profile_uuid,
		team_id,
		name,
		'darwin' as platform,
		identifier,
		checksum,
		created_at,
		updated_at
	FROM
		mdm_apple_configuration_profiles
	WHERE
		team_id = ? AND
		identifier NOT IN (?)

	UNION

	SELECT
		profile_uuid,
		team_id,
		name,
		'windows' as platform,
		'' as identifier,
		'' as checksum,
		created_at,
		updated_at
	FROM
		mdm_windows_configuration_profiles
	WHERE
		team_id = ?
) as combined_profiles
`

	var globalOrTeamID uint
	if teamID != nil {
		globalOrTeamID = *teamID
	}

	fleetIdentsMap := mobileconfig.FleetPayloadIdentifiers()
	fleetIdentifiers := make([]string, 0, len(fleetIdentsMap))
	for k := range fleetIdentsMap {
		fleetIdentifiers = append(fleetIdentifiers, k)
	}

	args := []any{globalOrTeamID, fleetIdentifiers, globalOrTeamID}
	stmt, args := appendListOptionsWithCursorToSQL(selectStmt, args, &opt)

	stmt, args, err := sqlx.In(stmt, args...)
	if err != nil {
		return nil, nil, ctxerr.Wrap(ctx, err, "sqlx.In ListMDMConfigProfiles")
	}

	if err := sqlx.SelectContext(ctx, ds.reader(ctx), &profs, stmt, args...); err != nil {
		return nil, nil, ctxerr.Wrap(ctx, err, "select profiles")
	}

	var metaData *fleet.PaginationMetadata
	if opt.IncludeMetadata {
		metaData = &fleet.PaginationMetadata{HasPreviousResults: opt.Page > 0}
		if len(profs) > int(opt.PerPage) {
			metaData.HasNextResults = true
			profs = profs[:len(profs)-1]
		}
	}
	return profs, metaData, nil
}

// Note that team ID 0 is used for profiles that apply to hosts in no team
// (i.e. pass 0 in that case as part of the teamIDs slice). Only one of the
// slice arguments can have values.
func (ds *Datastore) BulkSetPendingMDMHostProfiles(
	// TODO(mna): switch to all uuids, distinguish using uuid prefix?
	ctx context.Context,
	hostIDs, teamIDs, profileIDs []uint,
	profileUUIDs, hostUUIDs []string,
) error {
	var countArgs int
	if len(hostIDs) > 0 {
		countArgs++
	}
	if len(teamIDs) > 0 {
		countArgs++
	}
	if len(profileIDs) > 0 {
		countArgs++
	}
	if len(profileUUIDs) > 0 {
		countArgs++
	}
	if len(hostUUIDs) > 0 {
		countArgs++
	}
	if countArgs > 1 {
		return errors.New("only one of hostIDs, teamIDs, profileIDs, profileUUIDs or hostUUIDs can be provided")
	}
	if countArgs == 0 {
		return nil
	}

	var (
		hosts    []fleet.Host
		args     []any
		uuidStmt string
	)

	switch {
	case len(hostUUIDs) > 0:
		// TODO: if a very large number (~65K) of uuids was provided, could
		// result in too many placeholders (not an immediate concern).
		uuidStmt = `SELECT uuid, platform FROM hosts WHERE uuid IN (?)`
		args = append(args, hostUUIDs)

	case len(hostIDs) > 0:
		// TODO: if a very large number (~65K) of uuids was provided, could
		// result in too many placeholders (not an immediate concern).
		uuidStmt = `SELECT uuid, platform FROM hosts WHERE id IN (?)`
		args = append(args, hostIDs)

	case len(teamIDs) > 0:
		// TODO: if a very large number (~65K) of team IDs was provided, could
		// result in too many placeholders (not an immediate concern).
		uuidStmt = `SELECT uuid, platform FROM hosts WHERE `
		if len(teamIDs) == 1 && teamIDs[0] == 0 {
			uuidStmt += `team_id IS NULL`
		} else {
			uuidStmt += `team_id IN (?)`
			args = append(args, teamIDs)
			for _, tmID := range teamIDs {
				if tmID == 0 {
					uuidStmt += ` OR team_id IS NULL`
					break
				}
			}
		}

	case len(profileIDs) > 0:
		// TODO: if a very large number (~65K) of profile IDs was provided, could
		// result in too many placeholders (not an immediate concern).
		uuidStmt = `
SELECT DISTINCT h.uuid, h.platform
FROM hosts h
JOIN mdm_apple_configuration_profiles macp
	ON h.team_id = macp.team_id OR (h.team_id IS NULL AND macp.team_id = 0)
WHERE
	macp.profile_id IN (?) AND h.platform = 'darwin'`
		args = append(args, profileIDs)

	case len(profileUUIDs) > 0:
		// TODO: if a very large number (~65K) of profile IDs was provided, could
		// result in too many placeholders (not an immediate concern).
		uuidStmt = `
SELECT DISTINCT h.uuid, h.platform
FROM hosts h
JOIN mdm_windows_configuration_profiles mawp
	ON h.team_id = mawp.team_id OR (h.team_id IS NULL AND mawp.team_id = 0)
WHERE
	mawp.profile_uuid IN (?) AND h.platform = 'windows'`
		args = append(args, profileUUIDs)

	}

	return ds.withTx(ctx, func(tx sqlx.ExtContext) error {
		// TODO: this could be optimized to avoid querying for platform when
		// profileIDs or profileUUIDs are provided.
		if len(hosts) == 0 {
			uuidStmt, args, err := sqlx.In(uuidStmt, args...)
			if err != nil {
				return ctxerr.Wrap(ctx, err, "prepare query to load host UUIDs")
			}
			if err := sqlx.SelectContext(ctx, tx, &hosts, uuidStmt, args...); err != nil {
				return ctxerr.Wrap(ctx, err, "execute query to load host UUIDs")
			}
		}

		var macHosts []string
		var winHosts []string
		for _, h := range hosts {
			switch h.Platform {
			case "darwin":
				macHosts = append(macHosts, h.UUID)
			case "windows":
				winHosts = append(winHosts, h.UUID)
			default:
				level.Debug(ds.logger).Log(
					"msg", "tried to set profile status for a host with unsupported platform",
					"platform", h.Platform,
					"host_uuid", h.UUID,
				)
			}
		}

		if err := ds.bulkSetPendingMDMAppleHostProfilesDB(ctx, tx, macHosts); err != nil {
			return ctxerr.Wrap(ctx, err, "bulk set pending apple host profiles")
		}

		if err := ds.bulkSetPendingMDMWindowsHostProfilesDB(ctx, tx, winHosts); err != nil {
			return ctxerr.Wrap(ctx, err, "bulk set pending windows host profiles")
		}

		return nil
	})
}
