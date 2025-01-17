import React from "react";
import ReactTooltip from "react-tooltip";
import { uniqueId } from "lodash";

import Icon from "components/Icon";
import TextCell from "components/TableContainer/DataTable/TextCell";
import {
  FLEET_FILEVAULT_PROFILE_DISPLAY_NAME,
  ProfileOperationType,
} from "interfaces/mdm";

import {
  isMdmProfileStatus,
  OsSettingsTableStatusValue,
} from "../OSSettingsTableConfig";
import TooltipContent from "./components/Tooltip/TooltipContent";
import {
  PROFILE_DISPLAY_CONFIG,
  ProfileDisplayOption,
  WINDOWS_DISK_ENCRYPTION_DISPLAY_CONFIG,
} from "./helpers";

const baseClass = "os-setting-status-cell";

interface IOSSettingStatusCellProps {
  status: OsSettingsTableStatusValue;
  operationType: ProfileOperationType | null;
  profileName: string;
}

const OSSettingStatusCell = ({
  status,
  operationType,
  profileName = "",
}: IOSSettingStatusCellProps) => {
  let displayOption: ProfileDisplayOption = null;

  // windows hosts do not have an operation type at the moment and their display options are
  // different than mac hosts.
  if (!operationType && isMdmProfileStatus(status)) {
    displayOption = WINDOWS_DISK_ENCRYPTION_DISPLAY_CONFIG[status];
  }

  if (operationType) {
    displayOption = PROFILE_DISPLAY_CONFIG[operationType]?.[status];
  }

  const isDeviceUser = window.location.pathname
    .toLowerCase()
    .includes("/device/");

  const isDiskEncryptionProfile =
    profileName === FLEET_FILEVAULT_PROFILE_DISPLAY_NAME;

  if (displayOption) {
    const { statusText, iconName, tooltip } = displayOption;
    const tooltipId = uniqueId();
    return (
      <span className={baseClass}>
        <Icon name={iconName} />
        {tooltip ? (
          <>
            <span
              className="tooltip tooltip__tooltip-icon"
              data-tip
              data-for={tooltipId}
              data-tip-disable={false}
            >
              {statusText}
            </span>
            <ReactTooltip
              place="top"
              effect="solid"
              backgroundColor="#3e4771"
              id={tooltipId}
              data-html
            >
              <span className="tooltip__tooltip-text">
                {status !== "action_required" ? (
                  <TooltipContent
                    innerContent={tooltip}
                    innerProps={{
                      isDiskEncryptionProfile,
                    }}
                  />
                ) : (
                  <TooltipContent
                    innerContent={tooltip}
                    innerProps={{ isDeviceUser, profileName }}
                  />
                )}
              </span>
            </ReactTooltip>
          </>
        ) : (
          statusText
        )}
      </span>
    );
  }
  // graceful error - this state should not be reached based on the API spec
  return <TextCell value="Unrecognized" />;
};
export default OSSettingStatusCell;
