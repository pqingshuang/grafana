///<reference path="../../../headers/common.d.ts" />

import config from 'app/core/config';
import _ from 'lodash';
import $ from 'jquery';
import coreModule from '../../core_module';
import {NavModel, NavModelItem}  from '../../nav_model_srv';

export class NavbarCtrl {
  model: NavModel;

  /** @ngInject */
  constructor(private $scope, private $rootScope, private contextSrv) {
    console.log(this.model);
  }

  showSearch() {
    this.$rootScope.appEvent('show-dash-search');
  }

  navItemClicked(navItem, evt) {
    if (navItem.clickHandler) {
      navItem.clickHandler();
      evt.preventDefault();
    }
  }
}

export function navbarDirective() {
  return {
    restrict: 'E',
    templateUrl: 'public/app/core/components/navbar/navbar.html',
    controller: NavbarCtrl,
    bindToController: true,
    controllerAs: 'ctrl',
    scope: {
      model: "=",
    },
    link: function(scope, elem) {
    }
  };
}

coreModule.directive('navbar', navbarDirective);
