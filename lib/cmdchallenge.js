/* eslint strict: ["error", "global"] */
'use strict';

var HOSTNAME = window.location.hostname.split('.');
var BASEURL = HOSTNAME.filter(function (i) {
  return !["oops", "12days"].includes(i);
}).join(".");
var OOPS_IMG = '<img src="img/emojis/1F92D.png" alt="" />';
var CMD_IMG = '<img src="img/cmdchallenge-round.png" alt="" />';
var XMAS_IMG = '<img src="img/emojis/1F384.png" alt="" />';
var BASEURLS = {
  CMD: '//' + BASEURL,
  OOPS: '//oops.' + BASEURL,
  XMAS: '//12days.' + BASEURL
};
var SITES = {
  CMD: "cmdchallenge",
  OOPS: "oops",
  XMAS: "12days"
};
var FLAVOR = ["oops", "12days"].includes(HOSTNAME[0]) ? HOSTNAME[0] : "cmdchallenge";
var SITE_LINKS = {
  CMD: '<a href="//' + BASEURLS.CMD + '">' + CMD_IMG + '</a>',
  OOPS: '<a href="//' + BASEURLS.OOPS + '">' + OOPS_IMG + '</a>',
  XMAS: '<a href="//' + BASEURLS.XMAS + '">' + XMAS_IMG + '</a>'
};
var CMD_URL = window.location.hostname.match(/local/) ? 'https://testing.cmdchallenge.com/r' : '/r';
var TAB_COMPLETION = FLAVOR === SITES.OOPS ? ['echo', 'read'] : ['find', 'echo', 'awk', 'sed', 'perl', 'wc', 'grep', 'cat', 'sort', 'cut', 'ls', 'tac', 'jq', 'paste', 'tr', 'rm', 'tail', 'comm', 'egrep'];
var STORAGE_CORRECT = 'correct_answers';
var INFO_STATUS = {
  incorrect: 'incorrect',
  correct: 'correct',
  error: 'error'
};
jQuery(function ($) {
  var term;
  var currentChallenge = null;
  var challenges = [];
  var retCode;
  var stepGen = /*#__PURE__*/regeneratorRuntime.mark(function stepGen(steps) {
    return regeneratorRuntime.wrap(function stepGen$(_context) {
      while (1) {
        switch (_context.prev = _context.next) {
          case 0:
            if (!true) {
              _context.next = 4;
              break;
            }

            return _context.delegateYield(steps, "t0", 2);

          case 2:
            _context.next = 0;
            break;

          case 4:
          case "end":
            return _context.stop();
        }
      }
    }, stepGen);
  });
  var errorEmoji = stepGen(['emojis/1F63F.png']);
  var incorrectEmoji = stepGen(['emojis/E282.png', 'emojis/1F645-200D-2640-FE0F.png', 'emojis/1F645-200D-2642-FE0F.png', 'emojis/1F940.png']);
  var correctEmojiBeg = stepGen(['emojis/1F471-200D-2640-FE0F.png', 'emojis/1F471-200D-2642-FE0F.png']);
  var correctEmojiInt = stepGen(['emojis/1F9D4.png', 'emojis/1F468-200D-1F9B1.png', 'emojis/1F468-200D-1F9B0.png', 'emojis/1F468-200D-1F33E.png', 'emojis/1F468-200D-1F52C.png', 'emojis/1F468-200D-1F373.png', 'emojis/1F468-200D-1F393.png']);
  var correctEmojiAdv = stepGen(['emojis/1F478.png', 'emojis/1F482.png', 'emojis/1F9DD.png', 'emojis/1F9DD-200D-2640-FE0F.png', 'emojis/1F680.png']);
  var correctEmojiOops = stepGen(['emojis/1F600.png', 'emojis/1F604.png', 'emojis/1F970.png', 'emojis/1F60D.png', 'emojis/1F929.png']);
  var correctEmojiXmas = stepGen(['emojis/1F936.png', 'emojis/1F385.png', 'emojis/1F36D.png', 'emojis/2603.png']);
  var cmReader = new commonmark.Parser();
  var cmWriter = new commonmark.HtmlRenderer();

  var htmlFromMarkdown = function htmlFromMarkdown(markdown) {
    var parsed = cmReader.parse(markdown);
    return cmWriter.render(parsed);
  };

  var termClear = function termClear() {
    if (currentChallenge && !less) {
      retCode = colorize('0', 'green');
    }
  };

  var getArrayFromStorage = function getArrayFromStorage(storageName) {
    var ids;

    try {
      ids = JSON.parse(localStorage.getItem(storageName));
    } catch (e) {
      ids = [];
    }

    if (ids === null) {
      ids = [];
    }

    return ids.filter(function (v, i, a) {
      return a.indexOf(v) === i;
    });
  };

  var addItemToStorage = function addItemToStorage(item, storageName, callback) {
    var jsonItems;
    var items = getArrayFromStorage(storageName);

    if (!items.includes(item)) {
      jsonItems = JSON.stringify(items.concat([item]));
      localStorage.setItem(storageName, jsonItems);
    }

    if (typeof callback === 'function') {
      callback();
    }
  };

  var checkForWin = function checkForWin() {
    if (uncompletedChallenges().length === 0) {
      switch (FLAVOR) {
        case SITES.OOPS:
          $('.title .won').html('🎉 Congrats, you completed the challenge! 🎉 Try <a href="' + BASEURLS.CMD + '">even more challenges!</a>').show();
          break;

        case SITES.CMD:
          $('.title .won').html('🎉 Congrats, you completed the challenge! 🎉').show();
          break;
      }

      return true;
    } else {
      $('.title .won').hide();
      return false;
    }
  };

  var colorize = function colorize(msg, color, effect) {
    if (!effect) {
      effect = '';
    }
    /*
      u — underline.
      s — strike.
      o — overline.
      i — italic.
      b — bold.
      g — glow (using css text-shadow).
    */


    return '[[' + effect + ';' + color + ';black]' + msg + ']';
  };

  var underlineCurrent = function underlineCurrent() {
    var slug = currentChallenge.slug;
    challenges.forEach(function (challenge) {
      if (slug == challenge.slug) {
        $('#' + challenge.slug).removeClass('active-challenge inactive-challenge').addClass('active-challenge');
        $('.img-container.' + challenge.slug).removeClass('active-badge inactive-badge').addClass('active-badge');
      } else {
        $('#' + challenge.slug).removeClass('active-challenge inactive-challenge').addClass('inactive-challenge');
        $('.img-container.' + challenge.slug).removeClass('active-badge inactive-badge').addClass('inactive-badge');
      }
    });
  };

  var activeChallenges = function activeChallenges() {
    // completed challenges + the first uncompleted challenge
    return completedChallenges().concat(uncompletedChallenges()[0] || []);
  };

  var updateRoutes = function updateRoutes(callback) {
    var routes = {};
    challenges.forEach(function (c) {
      var slug = c.slug;

      routes['/' + slug] = function () {
        currentChallenge = c;
        updateChallengeDesc();
        updateChallenges();
        checkForWin();
      };
    });

    routes['*'] = function () {
      currentChallenge = uncompletedChallenges()[0] || challenges[0];
      updateChallengeDesc();
      updateChallenges();
      checkForWin();
    };

    routie(routes);

    if (typeof callback === 'function') {
      callback();
    }
  };

  var updateChallenges = function updateChallenges(callback) {
    // update the badges
    $('div#badges').html('');
    activeChallenges().forEach(function (c) {
      var slug = c.slug;
      var dispTitle = c.disp_title;
      $('div#badges').append('<div tabindex=\'-1\' class=\'img-container ' + slug + '\'><a title=\'' + dispTitle + '\' id=\'badge_' + slug + '\' href=\'#/' + slug + '\'><img class=\'badge\' src=\'img/' + c.emoji + '.png\' alt=\'' + slug + '\'/></a></li>');
      $('a#badge_' + slug).on('click', function (e) {
        e.preventDefault();
        e.stopPropagation();
        term.focus();
        routie('/' + slug);
      });
    });
    displaySolution();
    underlineCurrent();
    $('#learn').html('');

    if (currentChallenge.learn) {
      displayLearn();
    }

    if (typeof callback === 'function') {
      callback();
    }
  };

  var getChallenges = function getChallenges(callback) {
    $.ajax({
      dataType: 'json',
      url: '/challenges/challenges.json',
      success: function success(resp) {
        if (typeof callback === 'function') {
          switch (FLAVOR) {
            case SITES.OOPS:
              callback(resp.filter(function (o) {
                return (o.tags || []).includes('oops');
              }));
              break;

            case SITES.XMAS:
              callback(resp.filter(function (o) {
                return (o.tags || []).includes('12days');
              }));
              break;

            case SITES.CMD:
              callback(resp.filter(function (o) {
                return !o.tags;
              }));
              break;
          }
        }
      },
      error: function error() {
        retCode = '☠️';
        updateInfoText('Unable to retrieve challenges :(', INFO_STATUS.error);
      }
    });
  };

  var sendCommand = function sendCommand(command, callback) {
    var data = {
      'cmd': command,
      'challenge_slug': currentChallenge.slug,
      'version': currentChallenge.version,
      'img': currentChallenge.img || 'cmd'
    };
    $.ajax({
      type: 'GET',
      url: CMD_URL,
      // dataType: 'json',
      async: true,
      // contentType: "application/json; charset=utf-8",
      data: data,
      success: function success(resp) {
        if (typeof callback === 'function') {
          callback(resp);
        }
      },
      error: function error(resp) {
        if (typeof callback === 'function') {
          var output = resp.responseText || 'Unknown Error :(';
          callback({
            output: output,
            correct: false,
            return_code: '☠️'
          });
        }
      }
    });
  };

  var clearChallengeOutput = function clearChallengeOutput() {
    $('#challenge-output').text('').hide();
  };

  var updateInfoText = function updateInfoText(msg, infoStatus) {
    $('#info-box .text').html(msg);

    switch (infoStatus) {
      case INFO_STATUS.correct:
        var index = challenges.indexOf(currentChallenge);
        var emojiFname;

        switch (FLAVOR) {
          case SITES.OOPS:
            emojiFname = correctEmojiOops.next().value;
            break;

          case SITES.CMD:
            if (index < 4) {
              emojiFname = correctEmojiBeg.next().value;
            } else if (index >= 4 && index < 20) {
              emojiFname = correctEmojiInt.next().value;
            } else {
              emojiFname = correctEmojiAdv.next().value;
            }

            break;

          case SITES.XMAS:
            emojiFname = correctEmojiXmas.next().value;
            break;
        }

        $('#info-box .gradient').removeClass('incorrect correct error').addClass('correct');
        $('#info-box .img').html('<img src=\'img/' + emojiFname + '\' alt=\'correct\' />');
        break;

      case INFO_STATUS.incorrect:
        $('#info-box .gradient').removeClass('incorrect correct error').addClass('incorrect');
        $('#info-box .img').html('<img src=\'img/' + incorrectEmoji.next().value + '\' alt=\'incorrect\' />');
        break;

      case INFO_STATUS.error:
        $('#info-box .gradient').removeClass('incorrect correct error').addClass('error');
        $('#info-box .img').html('<img src=\'img/' + errorEmoji.next().value + '\' alt=\'correct\' />');
        break;

      default:
        throw new Error('Invalid status: ' + infoStatus);
    }

    $('#info-box').show();
  };

  var updateChallengeDesc = function updateChallengeDesc() {
    var description = htmlFromMarkdown(currentChallenge.description);
    $('#challenge-desc .img-container').html('<img src=\'img/' + currentChallenge.emoji + '.png\' alt=\'' + currentChallenge.disp_title + '\' />');
    $('#challenge-desc .desc-container').html(description);
  };

  var updateChallengeOutput = function updateChallengeOutput(output) {
    var lines = output.split('\n');
    $('#challenge-output').text('');
    lines.forEach(function (line) {
      $('#challenge-output').append('<span>' + line + '</span>');
    });
    $('#challenge-output').show();
  };

  var uncompletedChallenges = function uncompletedChallenges() {
    var completed = getArrayFromStorage(STORAGE_CORRECT);
    return challenges.filter(function (o) {
      return !completed.includes(o.slug);
    });
  };

  var completedChallenges = function completedChallenges() {
    var completed = getArrayFromStorage(STORAGE_CORRECT);
    return challenges.filter(function (o) {
      return completed.includes(o.slug);
    });
  };

  var displayLearn = function displayLearn() {
    if (currentChallenge.disp_learn) {
      $('#chck2').prop('checked', true);
    } else {
      $('#chck2').prop('checked', false);
    }

    $('#learn').html(htmlFromMarkdown(currentChallenge.learn));
    $('#learn-box').show();
  };

  var displaySolution = function displaySolution() {
    $.ajax({
      dataType: 'json',
      url: '/s/solutions/' + currentChallenge.slug + '.json',
      success: function success(resp) {
        $('#solutions').html('').show();
        resp.cmds.forEach(function (cmd) {
          $('#solutions').append(escapeHtml(cmd) + '\n');
        });
        hljs.highlightBlock(document.getElementById('solutions'));
        $('#solutions-wrapper .last-updated').html('Solutions updated ' + dateDelta(resp.ts) + ' ago');
      },
      error: function error() {
        $('#solutions').html('').hide();
        $('#solutions-wrapper .last-updated').html('No solutions for this challenge yet');
      }
    });
  };

  var escapeHtml = function escapeHtml(unsafe) {
    return unsafe.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#039;').replace(/\n/g, '\n  ');
  };

  var dateDelta = function dateDelta(lastUpdated) {
    var curTime = new Date().getTime() / 1000;
    var delta = Math.round(curTime - lastUpdated);
    var minutesDelta = Math.floor(delta / 60);
    var secondsDelta = delta % 60;
    var timeDisp = '';

    if (minutesDelta) {
      if (minutesDelta === 1) {
        timeDisp += minutesDelta + ' minute ';
      } else {
        timeDisp += minutesDelta + ' minutes ';
      }
    }

    if (secondsDelta === 1) {
      timeDisp += ' and ' + secondsDelta + ' second';
    } else {
      timeDisp += ' and ' + secondsDelta + ' seconds';
    }

    return timeDisp;
  };

  var processResp = function processResp(resp) {
    if (resp.return_code === 0) {
      retCode = colorize(resp.return_code, 'green');
    } else {
      retCode = colorize(resp.return_code, 'red');
    }

    if (isNaN(resp.return_code)) {
      updateInfoText('Unable to process command - got response: ' + resp.output, INFO_STATUS.error);
    } else {
      updateChallengeOutput(resp.output);

      if (resp.correct) {
        addItemToStorage(resp.challenge_slug, STORAGE_CORRECT, function () {
          updateChallenges();
          currentChallenge = uncompletedChallenges()[0] || challenges[0];

          if (checkForWin()) {
            updateInfoText('Correct! You you completed all of the challenges, ' + 'but feel free to keep on going!', INFO_STATUS.correct);
          } else {
            updateInfoText('Correct! You have a new challenge!', INFO_STATUS.correct);
          }

          routie('/' + currentChallenge.slug);
        });
      } else {
        updateInfoText('Incorrect answer, try again', INFO_STATUS.incorrect);

        if (resp.test_errors && resp.test_errors.length > 0) {
          updateInfoText(resp.test_errors[0] + ' - try again', INFO_STATUS.incorrect);
        } else if (resp.rand_error) {
          updateInfoText('Test against random data failed - try again', INFO_STATUS.incorrect);
        }
      }
    }

    term.resume();
  }; // main
  // Setup for different site types


  switch (FLAVOR) {
    case SITES.OOPS:
      $('#header-img').html(OOPS_IMG);
      $('#header-text').html('Oops I deleted my bin/ dir :('); // $('#links ul').prepend('<li class="link">' + SITE_LINKS.XMAS + '</li>');

      $('#links ul').prepend('<li class="link">' + SITE_LINKS.CMD + '</li>');
      break;

    case SITES.XMAS:
      WebFont.load({
        custom: {
          families: ['Snowburst One', 'Princess Sofia'],
          urls: ['../css/fonts.css']
        },
        active: function active() {
          $('#header-img').html(XMAS_IMG);
          $('#header-text').addClass("snowburst");
          $('#header-text').html('12 Days of Shell');
          $('#links ul').prepend(SITE_LINKS.OOPS);
          $('#links ul').prepend(SITE_LINKS.CMD);
        }
      });
      break;

    case SITES.CMD:
      console.log("cmd!");
      $('#header-text').html('Command Challenge');
      $('#header-img').html(CMD_IMG); // $('#links ul').prepend(SITE_LINKS.XMAS);

      $('#links ul').prepend(SITE_LINKS.OOPS);
      break;
  }

  $('#header-img').show();
  $('#header-text').show();
  retCode = colorize('0', 'green'); // Prevent backspace from doing anything except
  // where we input text

  $(document).keydown(function (e) {
    var element = e.target.nodeName.toLowerCase();

    if (element != 'input' && element != 'textarea' || $(e.target).attr('readonly') || e.target.getAttribute('type') === 'checkbox') {
      if (e.keyCode === 8) {
        return false;
      }
    }
  });
  getChallenges(function (c) {
    challenges = c;
    $('#term-challenge').terminal(function (command, term) {
      // Remove beginning and trailing whitespace
      command = command.replace(/^\s+|\s+$/g, '');

      if (command !== '') {
        routie('/' + currentChallenge.slug);
        $('#info-box').hide();
        clearChallengeOutput();
        $('#chck1').prop('checked', false);

        if (/tail\s+-[Ff]/.test(command)) {
          updateInfoText('<code>tail -f</code> will wait for additional data ' + 'to be appended to the file, try removing the -f option', INFO_STATUS.incorrect);
        } else {
          term.pause();
          sendCommand(command, processResp);
        }
      } else {
        term.clear();
      }
    }, {
      greetings: '',
      name: 'cmdchallenge',
      height: 40,
      outputLimit: 1,
      convertLinks: true,
      linksNoReferrer: true,
      onBeforeCommand: function onBeforeCommand(command) {
        $('.terminal').hide();
        $('#term-spinner').show();
      },
      onAfterCommand: function onAfterCommand(command) {
        $('#term-spinner').hide();
        $('.terminal').show();
      },
      prompt: function prompt(callback) {
        callback('(' + retCode + ')> ');
      },
      completion: function completion(string, callback) {
        var completions = Array.prototype.concat(TAB_COMPLETION, currentChallenge.completions || []);
        return completions;
      },
      doubleTab: function doubleTab(completingString, completions, echoFn) {
        $('#completions').html(completions.join('<span style=\'color: grey;\'> | </span>')).show().fadeOut(5000);
      },
      onClear: termClear
    });
    term = $.terminal.active();
    updateRoutes();
  });
});