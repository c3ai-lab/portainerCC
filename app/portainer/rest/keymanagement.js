angular.module('portainer.app').factory('Keymanagement', [
  '$resource',
  'API_ENDPOINT_KEYMANAGEMENT',
  function KeymanagementFactory($resource, API_ENDPOINT_KEYMANAGEMENT) {
    'use strict';
    return $resource(
      API_ENDPOINT_KEYMANAGEMENT,
      {},
      {
        create: { method: 'POST', ignoreLoadingBar: true },
        query: { method: 'GET', isArray: true },
        remove: { method: 'DELETE', params: { id: '@id' } },
        queryMemberships: { method: 'GET', isArray: true, params: { id: '@id', entity: 'memberships' } },
      }
    );
  },
]);
