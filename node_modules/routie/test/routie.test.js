require('assert');

suite('routie', function() {

  teardown(function(done) {
    routie.removeAll();
    window.location.hash = '';
    setTimeout(done, 20);
  });

  test('noConflict', function() {
    var r = routie.noConflict();
    assert.equal(typeof window.routie, 'undefined');
    window.routie = r;
  });

  test('root route', function(done) {
    window.location.hash = '';
    //should be called right away since there is no hash
    routie('', function() {
      done();
    });
  });

  test('basic route', function(done) {
    routie('test', function() {
      done();
    });
    window.location.hash = 'test';
  });

  test('pass in object', function(done) {
    routie({
      'test2': function() {
        done();
      }
    });
    window.location.hash = 'test2';
  });

  test('calling the same route more than once', function(done) {
    var runCount = 0;
    routie('test8', function() {
      runCount++;
    });
    routie('test8', function() {
      assert.equal(runCount, 1);
      done();
    });
    window.location.hash = 'test8';
  });

  test('trigger hash', function(done) {
    routie('test3');
    setTimeout(function() {
      assert.equal(window.location.hash, '#test3');
      done();
    }, 20);
  });

  test('remove route', function(done) {
    var check = false;
    var test9 = function() {
      check = true;
    };
    routie('test9', test9);
    routie.remove('test9', test9);
    window.location.hash = 'test9';
    setTimeout(function() {
      assert.equal(check, false);
      done();
    }, 20);
  });

  test('remove all routes', function(done) {
    var check = false;
    var test9 = function() {
      check = true;
    };
    var test20 = function() {
      check = true;
    };
    routie('test9', test9);
    routie('test20', test20);
    routie.removeAll();
    window.location.hash = 'test9';
    setTimeout(function() {
      window.location.hash = 'test20';
    }, 20);
    setTimeout(function() {
      assert.equal(check, false);
      done();
    }, 40);
  });

  test('regex support', function(done) {

    routie('test4/:name', function(name) {
      assert.equal(name, 'bob');
      assert.equal(this.params.name, 'bob');
      done();
    });

    routie('test4/bob');
  });

  //test('route with dash', function(done) {
    //routie('test-:name', function(name) {
      //assert.equal(name, 'bob');
    //});
    //routie('test-bob');
  //});

  test('optional param support', function(done) {
    routie('test5/:name?', function(name) {
      assert.equal(name, undefined);
      assert.equal(this.params.name, undefined);
      done();
    });

    routie('test5/');
  });

  test('wildcard', function(done) {
    routie('test7/*', function() {
      done();
    });
    routie('test7/123/123asd');
  });

  test('catch all', function(done) {
    routie('*', function() {
      done();
    });
    routie('test6');
  });

  test('this set with data about the route', function(done) {
    routie('test', function() {
      assert.equal(this.path, 'test');
      done();
    });
    routie('test');
  });

  test('double fire bug', function(done) {
    var called = 0;
    routie({
      'splash1': function() {
        routie('splash2');
      },
      'splash2': function() {
        called++;
      }
    });
    routie('splash1');

    setTimeout(function() {
      assert.equal(called, 1);
      done();
    }, 100);
  });

  test('only first route is run', function(done) {
    var count = 0;
    routie({
      'test*': function() {
        count++;
      },
      'test10': function() {
        count++;
      }
    });
    routie('test10');
    setTimeout(function() {
      assert.equal(count, 1);
      done();
    }, 100);
  });

  test('fallback not called if something else matches', function(done) {
    var count = 0;
    routie({
      '': function() {
        //root
      },
      'test11': function() {
        count++;
      },
      '*': function() {
        count++;
      }
    });
    routie('test11');
    setTimeout(function() {
      assert.equal(count, 1);
      done();
    }, 100);
  });

  test('fallback called if nothing else matches', function(done) {
    var count = 0;
    routie({
      '': function() {
        //root
      },
      'test11': function() {
        count++;
      },
      '*': function() {
        count++;
      }
    });
    routie('test12');
    setTimeout(function() {
      assert.equal(count, 1);
      done();
    }, 100);
  });

  /*TODO
  test('route object passed as this', function(done) {
    routie('*', function() {
      console.log(this);
      assert.equal(this.route, 'test7');
      done();
    });
    routie('test7');
  });
  */

  suite('named routes', function() {
    test('allow for named routes', function() {
      routie('namedRoute name/', function() {});

      assert.equal(routie.lookup('namedRoute'), 'name/');
    });

    test('routes should still work the same', function(done) {

      routie('namedRoute url/name2/', done);
      routie('url/name2/');

    });

    test('allow for named routes with params', function() {
      routie({
        'namedRoute name2/:param': function() { }
      });

      assert.equal(routie.lookup('namedRoute', { param: 'test' }), 'name2/test');
    });

    test('allow for named routes with optional params', function() {
      routie({
        'namedRoute name2/:param?': function() { }
      });

      assert.equal(routie.lookup('namedRoute'), 'name2/');
    });

    test('allow for named routes with optional params', function() {
      routie({
        'namedRoute name2/:param?': function() { }
      });

      assert.equal(routie.lookup('namedRoute', { param: 'test' }), 'name2/test');
    });

    test('error if param not passed in', function() {
      routie({
        'namedRoute name2/:param': function() {
        }
      });

      assert.throws(function() {
        routie.lookup('namedRoute');
      });
    });

    test('this contains named route', function(done) {
      routie('namedRoute test/:param', function() {
        assert.equal(this.name, 'namedRoute');
        assert.equal(this.params.param, 'bob');
        done();
      });
      routie('test/bob');
    });

  });

  suite('navigate', function() {

    test('call routie.navigate to change hash', function(done) {
      //same as routie('nav-test');
      routie.navigate('nav-test');
      setTimeout(function() {
        assert.equal(window.location.hash, '#nav-test');
        done();
      }, 20);
    });

    test('pass in {silent: true} to not trigger route', function(done) {

      var called = 0;

      routie('silent-test', function() {
        called++;
      });

      routie.navigate('silent-test', { silent: true });

      setTimeout(function() {
        assert.equal(called, 0);
        assert.equal(window.location.hash, '#silent-test');
        done();
      }, 20);
    });
  });

});
