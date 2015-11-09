var devhubApp = angular.module('devhubApp', []);

devhubApp.controller('ImageListCtrl', function ($scope, $http) {
  $http.get('/api/images').success(function (data) {
    $scope.images = data.images;
  });
});

devhubApp.controller('BootscriptListCtrl', function ($scope, $http) {
  $http.get('/api/bootscripts').success(function (data) {
    $scope.bootscripts = data.bootscripts;
  });
});
