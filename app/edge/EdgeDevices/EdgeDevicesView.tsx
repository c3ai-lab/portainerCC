import { useState } from 'react';

import { PageHeader } from '@/portainer/components/PageHeader';
import { useSettings } from '@/portainer/settings/queries';
import { useGroups } from '@/portainer/environment-groups/queries';
import { r2a } from '@/react-tools/react2angular';

import { EdgeDevicesDatatableContainer } from '../devices/components/EdgeDevicesDatatable/EdgeDevicesDatatableContainer';

import { Loader } from './Loader';

export function EdgeDevicesView() {
  const [loadingMessage, setLoadingMessage] = useState('');

  const settingsQuery = useSettings();
  const groupsQuery = useGroups();

  if (!settingsQuery.data || !groupsQuery.data) {
    return null;
  }

  const settings = settingsQuery.data;

  return (
    <>
      <PageHeader
        title="Edge Devices"
        reload
        breadcrumbs={[{ label: 'EdgeDevices' }]}
      />

      {loadingMessage ? (
        <Loader message={loadingMessage} />
      ) : (
        <EdgeDevicesDatatableContainer
          setLoadingMessage={setLoadingMessage}
          isFdoEnabled={
            settings.EnableEdgeComputeFeatures &&
            settings.fdoConfiguration.enabled
          }
          showWaitingRoomLink={
            process.env.PORTAINER_EDITION === 'BE' &&
            settings.EnableEdgeComputeFeatures &&
            !settings.TrustOnFirstConnect
          }
          isOpenAmtEnabled={
            settings.EnableEdgeComputeFeatures &&
            settings.openAMTConfiguration.enabled
          }
          mpsServer={settings.openAMTConfiguration.mpsServer}
          groups={groupsQuery.data}
          storageKey="edgeDevices"
        />
      )}
    </>
  );
}

export const EdgeDevicesViewAngular = r2a(EdgeDevicesView, []);
