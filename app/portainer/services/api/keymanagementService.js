
angular.module('portainer.app').factory('KeymanagementService', [
  '$q',
  'Keymanagement',
  function KeymanagementServiceFactory($q, Keymanagement) {
    'use strict';
    var service = {};

    service.generateKey = function (type, description, teamIds) {
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

      return deferred.promise;
    }

    service.updateTeams = function (id, teamIds) {
      var deferred = $q.defer();
      var payload = {
        teamIds: teamIds,
      };

      Keymanagement.update({ id: id }, payload)
        .$promise.then(function success(data) {
          deferred.resolve(data);
        }).catch(function error(err) {
          deferred.reject({ msg: 'Unable to update key', err: err })
        })

      return deferred.promise;
    };

    service.getKeys = function (type) {
      console.log("TODO: SET TYPE ----" + type)
      var deferred = $q.defer();
      Keymanagement.query({})
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
