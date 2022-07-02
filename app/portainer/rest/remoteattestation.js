angular.module('portainer.app').factory('Remoteattestation', [
  '$resource',
  'API_ENDPOINT_REMOTEATTESTATION',
  function KeymanagementFactory($resource, API_ENDPOINT_REMOTEATTESTATION) {
    'use strict';
    return $resource(
      API_ENDPOINT_REMOTEATTESTATION,
      {},
      {
        query: {
          method: 'GET', isArray: true
        },
      }
    );
  },
]);
