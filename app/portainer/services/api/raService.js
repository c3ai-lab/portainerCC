
angular.module('portainer.app').factory('RaService', [
  '$q',
  'Remoteattestation',
  function RaServiceFactory($q, Remoteattestation) {
    'use strict';
    var service = {};

    service.getImages = function () {
      var deferred = $q.defer();
      Remoteattestation.query()
        .$promise.then(function success(data) {
          deferred.resolve(data);
        }).catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve ra image list', err: err })
        });
      return deferred.promise;
    }

    return service;
  },
]);
