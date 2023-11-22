package tables

import (
	"database/sql"
	"fmt"
)

func init() {
	MigrationClient.AddMigration(Up_20231122155427, Down_20231122155427)
}

func Up_20231122155427(tx *sql.Tx) error {
	// update the windows profiles tables to use a 37-char uuid column for
	// the 'W' prefix.
	_, err := tx.Exec(`
ALTER TABLE host_mdm_windows_profiles
	CHANGE COLUMN profile_uuid profile_uuid VARCHAR(37) NOT NULL DEFAULT ''
`)
	if err != nil {
		return fmt.Errorf("failed to alter host_mdm_windows_profiles table: %w", err)
	}
	_, err = tx.Exec(`
ALTER TABLE mdm_windows_configuration_profiles
	CHANGE COLUMN profile_uuid profile_uuid VARCHAR(37) NOT NULL DEFAULT ''
`)
	if err != nil {
		return fmt.Errorf("failed to alter mdm_windows_configuration_profiles table: %w", err)
	}
	_, err = tx.Exec(`
UPDATE
	mdm_windows_configuration_profiles
SET
	profile_uuid = CONCAT('W', profile_uuid)
`)
	if err != nil {
		return fmt.Errorf("failed to update mdm_windows_configuration_profiles table: %w", err)
	}
	_, err = tx.Exec(`
UPDATE
	host_mdm_windows_profiles
SET
	profile_uuid = CONCAT('W', profile_uuid)
`)
	if err != nil {
		return fmt.Errorf("failed to update host_mdm_windows_profiles table: %w", err)
	}

	// update the apple profiles table to add the profile_uuid column and
	// temporarily drop the primary key until we fill those uuids.
	_, err = tx.Exec(`
ALTER TABLE mdm_apple_configuration_profiles
	-- 37 and not 36 because the UUID will be prefixed with 'A' to indicate
	-- that it's an Apple profile.
	ADD COLUMN profile_uuid VARCHAR(37) NOT NULL DEFAULT '',
	-- auto-increment column must have an index, so we create one before
	-- dropping the primary key to make it profile_uuid later.
	ADD UNIQUE KEY idx_mdm_apple_config_prof_id (profile_id),
	DROP PRIMARY KEY
`)
	if err != nil {
		return fmt.Errorf("failed to alter mdm_apple_configuration_profiles table: %w", err)
	}

	// generate the uuids for the apple profiles table
	_, err = tx.Exec(`
UPDATE
	mdm_apple_configuration_profiles
SET
	profile_uuid = CONCAT('A', uuid())
`)
	if err != nil {
		return fmt.Errorf("failed to update mdm_apple_configuration_profiles table: %w", err)
	}

	// set the profile uuid as the new primary key
	_, err = tx.Exec(`
ALTER TABLE mdm_apple_configuration_profiles
	ADD PRIMARY KEY (profile_uuid)`)
	if err != nil {
		return fmt.Errorf("failed to set primary key of mdm_apple_configuration_profiles table: %w", err)
	}

	// add the profile_uuid column to the host apple profiles table, keeping the
	// old id for now. Cannot be set as primary key yet as it may have duplicates
	// until we generate the uuids.
	_, err = tx.Exec(`
ALTER TABLE host_mdm_apple_profiles
	DROP PRIMARY KEY,
	ADD COLUMN profile_uuid VARCHAR(37) NOT NULL DEFAULT ''
`)
	if err != nil {
		return fmt.Errorf("failed to alter host_mdm_apple_profiles table: %w", err)
	}

	// update the apple host profiles table's profile_uuid based on its profile_id
	_, err = tx.Exec(`
UPDATE
	host_mdm_apple_profiles
SET
	profile_uuid = COALESCE((
		SELECT
			macp.profile_uuid
		FROM
			mdm_apple_configuration_profiles macp
		WHERE
			host_mdm_apple_profiles.profile_id = macp.profile_id
	), CONCAT('A', uuid()))
`)
	if err != nil {
		return fmt.Errorf("failed to update host_mdm_apple_profiles table: %w", err)
	}

	// drop the now unused profile_id column from the host apple profiles table
	_, err = tx.Exec(`ALTER TABLE host_mdm_apple_profiles
		ADD PRIMARY KEY (host_uuid, profile_uuid),
		DROP COLUMN profile_id`)
	if err != nil {
		return fmt.Errorf("failed to drop column from host_mdm_apple_profiles table: %w", err)
	}

	return nil
}

func Down_20231122155427(tx *sql.Tx) error {
	return nil
}
