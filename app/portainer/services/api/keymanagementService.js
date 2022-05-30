
angular.module('portainer.app').factory('KeymanagementService', [
  '$q',
  'Keymanagement',
  function KeymanagementServiceFactory($q, Keymanagement) {
    'use strict';
    var service = {};

    service.generateKey = function (type, description, teamIds) {
      console.log("hallo")
      var deferred = $q.defer();
      var payload = {
        type: type,
        description: description,
        teamIds: teamIds
      }

      Keymanagement.create({}, payload)
        .$promise.then(function success(data) {
          deferred.resolve(data);
        }).catch(function error(err) {
          deferred.reject({ msg: 'Unable to generate key', err: err })
        })

      return deferred.$promise;
    }

    service.getKeys = function (type) {
      console.log("TODO: SET TYPE ----" + type)
      var deferred = $q.defer();
      Keymanagement.query()
        .$promise.then(function success(data) {
          deferred.resolve(data);
        }).catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve keys', err: err })
        });
      return deferred.promise;
    }

    return service;
  },
]);
