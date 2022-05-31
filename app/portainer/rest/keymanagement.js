angular.module('portainer.app').factory('Keymanagement', [
  '$resource',
  'API_ENDPOINT_KEYMANAGEMENT',
  function KeymanagementFactory($resource, API_ENDPOINT_KEYMANAGEMENT) {
    'use strict';
    return $resource(
      API_ENDPOINT_KEYMANAGEMENT + '/:id/',
      {},
      {
        create: { method: 'POST', ignoreLoadingBar: true },
        query: { method: 'GET', isArray: true },
        update: { method: 'PUT', params: { id: '@id' } },
      }
    );
  },
]);
