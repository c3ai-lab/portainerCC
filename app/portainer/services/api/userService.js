import _ from 'lodash-es';

import axios from '@/portainer/services/axios';

const BASE_URL = '/users';

import PortainerError from '@/portainer/error';
import { filterNonAdministratorUsers } from '@/portainer/helpers/userHelper';
import { UserViewModel, UserTokenModel } from '../../models/user';
import { TeamMembershipModel } from '../../models/teamMembership';

export async function getUsers(includeAdministrators) {
  try {
    let { data } = await axios.get(BASE_URL);

    const users = data.map((user) => new UserViewModel(user));

    if (includeAdministrators) {
      return users;
    }

    return filterNonAdministratorUsers(users);
  } catch (e) {
    let err = e;
    if (err.isAxiosError) {
      err = new Error(e.response.data.message);
    }

    throw new PortainerError('Unable to retrieve users', err);
  }
}

export async function getUser(id) {
  try {
    const { data: user } = await axios.get(`${BASE_URL}/${id}`);

    return new UserViewModel(user);
  } catch (e) {
    let err = e;
    if (err.isAxiosError) {
      err = new Error(e.response.data.message);
    }

    throw new PortainerError('Unable to retrieve user details', err);
  }
}

/* @ngInject */
export function UserService($q, Users, TeamService, TeamMembershipService) {
  'use strict';
  var service = {};

  service.users = getUsers;

  service.user = getUser;

  service.createUser = function (username, password, role, teamIds) {
    var deferred = $q.defer();

    var payload = {
      username: username,
      password: password,
      role: role,
    };

    Users.create({}, payload)
      .$promise.then(function success(data) {
        var userId = data.Id;
        var teamMembershipQueries = [];
        angular.forEach(teamIds, function (teamId) {
          teamMembershipQueries.push(TeamMembershipService.createMembership(userId, teamId, 2));
        });
        $q.all(teamMembershipQueries).then(function success() {
          deferred.resolve();
        });
      })
      .catch(function error(err) {
        deferred.reject({ msg: 'Unable to create user', err: err });
      });

    return deferred.promise;
  };

  service.deleteUser = function (id) {
    return Users.remove({ id: id }).$promise;
  };

  service.updateUser = function (id, { password, role, username }) {
    return Users.update({ id }, { password, role, username }).$promise;
  };

  service.updateUserPassword = function (id, currentPassword, newPassword) {
    var payload = {
      Password: currentPassword,
      NewPassword: newPassword,
    };

    return Users.updatePassword({ id: id }, payload).$promise;
  };

  service.updateUserTheme = function (id, userTheme) {
    return Users.updateTheme({ id }, { userTheme }).$promise;
  };

  service.userMemberships = function (id) {
    var deferred = $q.defer();

    Users.queryMemberships({ id: id })
      .$promise.then(function success(data) {
        var memberships = data.map(function (item) {
          return new TeamMembershipModel(item);
        });
        deferred.resolve(memberships);
      })
      .catch(function error(err) {
        deferred.reject({ msg: 'Unable to retrieve user memberships', err: err });
      });

    return deferred.promise;
  };

  service.userLeadingTeams = function (id) {
    var deferred = $q.defer();

    $q.all({
      teams: TeamService.teams(),
      memberships: service.userMemberships(id),
    })
      .then(function success(data) {
        var memberships = data.memberships;
        var teams = data.teams.filter(function (team) {
          var membership = _.find(memberships, { TeamId: team.Id });
          if (membership && membership.Role === 1) {
            return team;
          }
        });
        deferred.resolve(teams);
      })
      .catch(function error(err) {
        deferred.reject({ msg: 'Unable to retrieve user teams', err: err });
      });

    return deferred.promise;
  };

  service.getAccessTokens = function (id) {
    var deferred = $q.defer();

    Users.getAccessTokens({ id: id })
      .$promise.then(function success(data) {
        var userTokens = data.map(function (item) {
          return new UserTokenModel(item);
        });
        deferred.resolve(userTokens);
      })
      .catch(function error(err) {
        deferred.reject({ msg: 'Unable to retrieve user tokens', err: err });
      });

    return deferred.promise;
  };

  service.deleteAccessToken = function (id, tokenId) {
    return Users.deleteAccessToken({ id: id, tokenId: tokenId }).$promise;
  };

  service.initAdministrator = function (username, password) {
    return Users.initAdminUser({ Username: username, Password: password }).$promise;
  };

  service.administratorExists = function () {
    var deferred = $q.defer();

    Users.checkAdminUser({})
      .$promise.then(function success() {
        deferred.resolve(true);
      })
      .catch(function error(err) {
        if (err.status === 404) {
          deferred.resolve(false);
        }
        deferred.reject({ msg: 'Unable to verify administrator account existence', err: err });
      });

    return deferred.promise;
  };

  return service;
}
