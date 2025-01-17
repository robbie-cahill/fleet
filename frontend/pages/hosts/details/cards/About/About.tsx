import React from "react";

import ReactTooltip from "react-tooltip";
import { HumanTimeDiffWithFleetLaunchCutoff } from "components/HumanTimeDiffWithDateTip";
import TooltipWrapper from "components/TooltipWrapper";
import CustomLink from "components/CustomLink";

import { IHostMdmData, IMunkiData, IDeviceUser } from "interfaces/host";
import {
  DEFAULT_EMPTY_CELL_VALUE,
  MDM_STATUS_TOOLTIP,
} from "utilities/constants";

interface IAboutProps {
  aboutData: { [key: string]: any };
  deviceMapping?: IDeviceUser[];
  munki?: IMunkiData | null;
  mdm?: IHostMdmData;
}

const About = ({
  aboutData,
  deviceMapping,
  munki,
  mdm,
}: IAboutProps): JSX.Element => {
  const renderPublicIp = () => {
    if (aboutData.public_ip !== DEFAULT_EMPTY_CELL_VALUE) {
      return aboutData.public_ip;
    }
    return (
      <>
        <span
          className="text-cell text-muted tooltip"
          data-tip
          data-for={"public-ip-tooltip"}
        >
          {aboutData.public_ip}
        </span>
        <ReactTooltip
          place="bottom"
          effect="solid"
          backgroundColor="#3e4771"
          id={"public-ip-tooltip"}
          data-html
          clickable
          delayHide={200} // need delay set to hover using clickable
        >
          Public IP address could not be
          <br /> determined.{" "}
          <CustomLink
            url="https://fleetdm.com/docs/deploying/configuration#public-i-ps-of-devices"
            text="Learn more"
            newTab
            iconColor="core-fleet-white"
          />
        </ReactTooltip>
      </>
    );
  };

  const renderSerialAndIPs = () => {
    return (
      <>
        <div className="info-grid__block">
          <span className="info-grid__header">Serial number</span>
          <span className="info-grid__data">{aboutData.hardware_serial}</span>
        </div>
        <div className="info-grid__block">
          <span className="info-grid__header">Private IP address</span>
          <span className="info-grid__data">{aboutData.primary_ip}</span>
        </div>
        <div className="info-grid__block">
          <span className="info-grid__header">Public IP address</span>
          <span className="info-grid__data">{renderPublicIp()}</span>
        </div>
      </>
    );
  };

  const renderMunkiData = () => {
    return munki ? (
      <>
        <div className="info-grid__block">
          <span className="info-grid__header">Munki version</span>
          <span className="info-grid__data">
            {munki.version || DEFAULT_EMPTY_CELL_VALUE}
          </span>
        </div>
      </>
    ) : null;
  };

  const renderMdmData = () => {
    if (!mdm?.enrollment_status) {
      return null;
    }
    return (
      <>
        <div className="info-grid__block">
          <span className="info-grid__header">MDM status</span>
          <span className="info-grid__data">
            <TooltipWrapper
              tipContent={MDM_STATUS_TOOLTIP[mdm.enrollment_status]}
            >
              {mdm.enrollment_status}
            </TooltipWrapper>
          </span>
        </div>
        <div className="info-grid__block">
          <span className="info-grid__header">MDM server URL</span>
          <span className="info-grid__data">
            {mdm.server_url || DEFAULT_EMPTY_CELL_VALUE}
          </span>
        </div>
      </>
    );
  };

  const renderDeviceUser = () => {
    if (!deviceMapping) {
      return null;
    }

    const numUsers = deviceMapping.length;
    const tooltipText = deviceMapping.map((d) => (
      <span key={Math.random().toString().slice(2)}>
        {d.email}
        <br />
      </span>
    ));

    return (
      <div className="info-grid__block">
        <span className="info-grid__header">Used by</span>
        <span className="info-grid__data">
          {numUsers > 1 ? (
            <>
              <span data-tip data-for="device_mapping" className="tooltip">
                {`${numUsers} users`}
              </span>
              <ReactTooltip
                effect="solid"
                backgroundColor="#3e4771"
                id="device_mapping"
                data-html
              >
                <span className={`tooltip__tooltip-text`}>{tooltipText}</span>
              </ReactTooltip>
            </>
          ) : (
            deviceMapping[0].email || DEFAULT_EMPTY_CELL_VALUE
          )}
        </span>
      </div>
    );
  };

  const renderGeolocation = () => {
    const geolocation = aboutData.geolocation;

    if (!geolocation) {
      return null;
    }

    const location = [geolocation?.city_name, geolocation?.country_iso]
      .filter(Boolean)
      .join(", ");
    return (
      <div className="info-grid__block">
        <span className="info-grid__header">Location</span>
        <span className="info-grid__data">{location}</span>
      </div>
    );
  };

  const renderBattery = () => {
    if (
      aboutData.batteries === null ||
      typeof aboutData.batteries !== "object"
    ) {
      return null;
    }
    return (
      <div className="info-grid__block">
        <span className="info-grid__header">Battery condition</span>
        <span className="info-grid__data">
          {aboutData.batteries?.[0]?.health}
        </span>
      </div>
    );
  };

  return (
    <div className="section about">
      <p className="section__header">About</p>
      <div className="info-grid">
        <div className="info-grid__block">
          <span className="info-grid__header">Added to Fleet</span>
          <span className="info-grid__data">
            <HumanTimeDiffWithFleetLaunchCutoff
              timeString={aboutData.last_enrolled_at ?? "Unavailable"}
            />
          </span>
        </div>
        <div className="info-grid__block">
          <span className="info-grid__header">Last restarted</span>
          <span className="info-grid__data">
            <HumanTimeDiffWithFleetLaunchCutoff
              timeString={aboutData.last_restarted_at}
            />
          </span>
        </div>
        <div className="info-grid__block">
          <span className="info-grid__header">Hardware model</span>
          <span className="info-grid__data">{aboutData.hardware_model}</span>
        </div>
        {renderSerialAndIPs()}
        {renderMunkiData()}
        {renderMdmData()}
        {renderDeviceUser()}
        {renderGeolocation()}
        {renderBattery()}
      </div>
    </div>
  );
};

export default About;
