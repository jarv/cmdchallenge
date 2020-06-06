/* eslint strict: ["error", "global"] */

'use strict';

const GLOBAL_VERSION = 2; // for cache busting
const CMD_URL = window.location.hostname == 'localhost' ? 'https://testing.cmdchallenge.com/r' : '/r';
const TAB_COMPLETION = ['find', 'echo', 'awk', 'sed', 'perl', 'wc', 'grep',
  'cat', 'sort', 'cut', 'ls', 'tac'];

const STORAGE_CORRECT = 'correct_answers';
const INFO_STATUS = {
  incorrect: 'incorrect',
  correct: 'correct',
  error: 'error',
};

jQuery(function($) {
  let term;
  let currentChallenge = null;
  let challenges = [];
  let retCode;

  // Emojis for info-box ---

  const stepGen = function* (steps) {
    while (true) yield* steps;
  };

  const errorEmoji=stepGen(['1F63F.png']);
  const incorrectEmoji=stepGen(
      ['E282.png', '1F645-200D-2640-FE0F.png', '1F645-200D-2642-FE0F.png']
  );
  const correctEmojiBeg=stepGen(
      ['1F471-200D-2640-FE0F.png', '1F471-200D-2642-FE0F.png']
  );

  const correctEmojiInt=stepGen(
      [
        '1F9D4.png', '1F468-200D-1F9B1.png', '1F468-200D-1F9B0.png',
        '1F468-200D-1F33E.png', '1F468-200D-1F52C.png',
        '1F468-200D-1F373.png', '1F468-200D-1F393.png',
      ]
  );

  const correctEmojiAdv=stepGen(
      ['1F478.png', '1F482.png', '1F9DD.png', '1F9DD-200D-2640-FE0F.png']
  );

  // ---------------------

  const cmReader = new commonmark.Parser();
  const cmWriter = new commonmark.HtmlRenderer();

  const htmlFromMarkdown = function(markdown) {
    const parsed = cmReader.parse(markdown);
    return cmWriter.render(parsed);
  };

  const termClear = function() {
    if (currentChallenge && !less) {
      retCode = colorize('0', 'green');
    }
  };

  const getArrayFromStorage = function(storageName) {
    let ids;

    try {
      ids = JSON.parse(localStorage.getItem(storageName));
    } catch (e) {
      ids = [];
    }
    if (ids === null) {
      ids = [];
    }
    return ids.filter((v, i, a) => a.indexOf(v) === i);
  };

  const addItemToStorage = function(item, storageName, callback) {
    let jsonItems;
    const items = getArrayFromStorage(storageName);

    if (! items.includes(item)) {
      jsonItems = JSON.stringify(items.concat([item]));
      localStorage.setItem(storageName, jsonItems);
    }

    if (typeof callback === 'function') {
      callback();
    }
  };

  const checkForWin = function() {
    if (uncompletedChallenges().length === 0) {
      $('.title .won').show();
      return true;
    } else {
      $('.title .won').hide();
      return false;
    }
  };

  const colorize = function(msg, color, effect) {
    if (! effect) {
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

  const underlineCurrent = function() {
    const slug = currentChallenge.slug;
    challenges.forEach(function(challenge) {
      if (slug == challenge.slug) {
        $('#' + challenge.slug).removeClass(
            'active-challenge inactive-challenge'
        ).addClass('active-challenge');
        $('.img-container.' + challenge.slug).removeClass(
            'active-badge inactive-badge'
        ).addClass('active-badge');
      } else {
        $('#' + challenge.slug).removeClass(
            'active-challenge inactive-challenge'
        ).addClass('inactive-challenge');
        $('.img-container.' + challenge.slug).removeClass(
            'active-badge inactive-badge'
        ).addClass('inactive-badge');
      }
    });
  };

  const activeChallenges = function() {
    // completed challenges + the first uncompleted challenge
    return completedChallenges().concat(uncompletedChallenges()[0] || []);
  };

  const updateRoutes = function(callback) {
    const routes = {};
    challenges.forEach(function(c) {
      const slug = c.slug;
      routes['/s/' + slug] = function() {
        currentChallenge = c;
        clearChallengeOutput();
        updateChallengeDesc();
        updateChallenges();
        displaySolution();
      };
      routes['/' + slug] = function() {
        currentChallenge = c;
        clearSolutions();
        updateChallengeDesc();
        updateChallenges();
        checkForWin();
      };
    });
    routes[''] = function() {
      currentChallenge = uncompletedChallenges()[0] || challenges[0];
      clearSolutions();
      updateChallengeDesc();
      updateChallenges();
      checkForWin();
    };
    routie(routes);
    if (typeof callback === 'function') {
      callback();
    }
  };

  const updateChallenges = function(callback) {
    // update the badges
    $('div#badges').html('');
    activeChallenges().forEach(function(c) {
      const slug = c.slug;
      const dispTitle = c.disp_title;
      $('div#badges').append(
          '<div tabindex=\'-1\' class=\'img-container ' +
            slug + '\'><a id=\'badge_' +
            slug + '\' href=\'#/' +
            slug + '\'><img class=\'badge\' src=\'img/' +
            slug + '.png\' alt=\'' +
            slug + '\'/><span class=\'tooltip\'>' +
            dispTitle +
            '</span></a></li>');
      $('a#badge_' + slug).on('click', function(e) {
        e.preventDefault();
        e.stopPropagation();
        $('#chck1').prop('checked', false);
        term.focus();
        routie('/' + slug);
      });
    });

    // update the solutions

    $('ul#challenges').html('');
    const completedSlugs = completedChallenges().map( (c) => c.slug);
    challenges.forEach(function(c) {
      const slug = c.slug;
      if (completedSlugs.indexOf(slug) !== -1) {
        $('ul#challenges').append(
            '<li tabindex=\'-1\'><img src=\'img/' + slug +
            '.png\' /><a class=\'enable\' id=\'' + slug +
            '\' href=\'#/' + slug + '\' title=\'' +
            slug + '\'>' + c.disp_title + '</a></li>');
        $('a#' + slug).on('click', function(e) {
          e.preventDefault();
          e.stopPropagation();
          routie('/s/' + slug);
          term.focus();
        });
      } else {
        // Blur and lock
        $('ul#challenges').append(
            '<li tabindex=\'-1\'><img src=\'img/lock.png\' />' +
            '<a class=\'disable\' id=\'' + slug +
            '\' href=\'#' +
            '\' title=\'' +
            slug + '\'>' + c.disp_title + '</a></li>');
      }
    });

    // highlight active challenge
    underlineCurrent();

    if (typeof callback === 'function') {
      callback();
    }
  };

  const getChallenges = function(callback) {
    $.ajax({
      dataType: 'json',
      url: '/challenges/challenges.json',
      success: function(resp) {
        if (typeof callback === 'function') {
          callback(resp);
        }
      },
      error: function() {
        retCode = '☠️';
        updateInfoText('Unable to retrieve challenges :(', INFO_STATUS.error);
      },
    });
  };

  const sendCommand = function(command, callback) {
    const data = {
      'cmd': command,
      'challenge_slug': currentChallenge.slug,
      'version': currentChallenge.version,
      'g_version': GLOBAL_VERSION,
    };
    $.ajax({
      type: 'GET',
      url: CMD_URL,
      // dataType: 'json',
      async: true,
      // contentType: "application/json; charset=utf-8",
      data: data,
      success: function(resp) {
        if (typeof callback === 'function') {
          callback(resp);
        }
      },
      error: function(resp) {
        if (typeof callback === 'function') {
          const output = resp.responseText || 'Unknown Error :(';
          callback({
            output: output,
            correct: false,
            return_code: '☠️',
          });
        }
      },
    });
  };

  const clearChallengeOutput = function() {
    $('#challenge-output').text('').hide();
  };

  const clearSolutions = function() {
    $('#solutions').html('');
    $('#solutions-wrapper').hide();
  };

  const showSolutions = function() {
    $('#solutions-wrapper').show();
  };

  const updateInfoText = function(msg, infoStatus) {
    $('#info-box .text').html(msg);
    switch (infoStatus) {
      case INFO_STATUS.correct:
        const index = challenges.indexOf(currentChallenge);
        let emojiFname;
        if (index < 4) {
          emojiFname = correctEmojiBeg.next().value;
        } else if (index >= 4 && index < 20) {
          emojiFname = correctEmojiInt.next().value;
        } else {
          emojiFname = correctEmojiAdv.next().value;
        }
        $('#info-box .gradient').removeClass(
            'incorrect correct error').addClass('correct');
        $('#info-box .img').html(
            '<img src=\'img/emojis/' + emojiFname +
            '\' alt=\'correct\' />'
        );
        break;
      case INFO_STATUS.incorrect:
        $('#info-box .gradient').removeClass(
            'incorrect correct error').addClass('incorrect');
        $('#info-box .img').html(
            '<img src=\'img/emojis/' + incorrectEmoji.next().value +
            '\' alt=\'incorrect\' />'
        );
        break;
      case INFO_STATUS.error:
        $('#info-box .gradient').removeClass(
            'incorrect correct error').addClass('error');
        $('#info-box .img').html(
            '<img src=\'img/emojis/' + errorEmoji.next().value +
            '\' alt=\'correct\' />'
        );
        break;
      default:
        throw new Error('Invalid status: ' + infoStatus);
    }
    $('#info-box').show();
  };

  const updateChallengeDesc = function() {
    const description = htmlFromMarkdown(currentChallenge.description);
    $('#challenge-desc .img-container').html('<img src=\'img/' +
        currentChallenge.slug +
        '.png\' alt=\'' + currentChallenge.disp_title + '\' />');
    $('#challenge-desc .desc-container').html(description);
  };

  const updateChallengeOutput = function(output) {
    const lines = output.split('\n');

    $('#challenge-output').text('');
    lines.forEach(function(line) {
      $('#challenge-output').append('<span>' + line + '</span>');
    });
    $('#challenge-output').show();
  };

  const uncompletedChallenges = function() {
    const completed = getArrayFromStorage(STORAGE_CORRECT);
    return challenges.filter((o) => !completed.includes(o.slug));
  };

  const completedChallenges = function() {
    const completed = getArrayFromStorage(STORAGE_CORRECT);
    return challenges.filter((o) => completed.includes(o.slug));
  };

  const displaySolution = function() {
    $('#info-box').hide();
    $.ajax({
      dataType: 'json',
      url: '/s/solutions/' + currentChallenge.slug + '.json',
      success: function(resp) {
        clearSolutions();
        resp.cmds.forEach(function(cmd) {
          $('#solutions').append(escapeHtml(cmd) + '\n');
        });
        hljs.highlightBlock(document.getElementById('solutions'));
        $('#solutions-wrapper .last-updated').html(
            'Solutions updated ' + dateDelta(resp.ts) + ' ago'
        );
        showSolutions();
      },
      error: function() {
        retCode = '☠️';
        updateInfoText('Unable to retrieve solutions :(', INFO_STATUS.error);
      },
    });
  };

  const escapeHtml = function(unsafe) {
    return unsafe
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;')
        .replace(/\n/g, '\n  ');
  };

  const dateDelta = function(lastUpdated) {
    const curTime = (new Date()).getTime() / 1000;
    const delta = Math.round(curTime - lastUpdated);
    const minutesDelta = Math.floor(delta / 60);
    const secondsDelta = delta % 60;
    let timeDisp = '';
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


  // main

  retCode = colorize('0', 'green');
  // Prevent backspace from doing anything except
  // where we input text
  $(document).keydown(function(e) {
    const element = e.target.nodeName.toLowerCase();
    if ((element != 'input' && element != 'textarea') ||
         $(e.target).attr('readonly') ||
         (e.target.getAttribute('type') ==='checkbox')) {
      if (e.keyCode === 8) {
        return false;
      }
    }
  });

  // Refocus the term when you click on solutions
  $('#chck1').change(
      function() {
        term.focus();
      }
  );

  getChallenges(function(c) {
    challenges = c;

    $('#term-challenge').terminal(function(command, term) {
      // Remove beginning and trailing whitespace
      command = command.replace(/^\s+|\s+$/g, '');
      if (command !== '') {
        routie('/' + currentChallenge.slug);
        $('#info-box').hide();
        clearChallengeOutput();
        $('#chck1').prop('checked', false);
        term.pause();
        sendCommand(command, function(resp) {
          if (resp.return_code === 0) {
            retCode = colorize(resp.return_code, 'green');
          } else {
            retCode = colorize(resp.return_code, 'red');
          }

          if (isNaN(resp.return_code)) {
            updateInfoText(
                'Unable to process command - got response: ' + resp.output,
                INFO_STATUS.error
            );
          } else {
            updateChallengeOutput(resp.output);
            if (resp.correct) {
              addItemToStorage(
                  resp.challenge_slug,
                  STORAGE_CORRECT,
                  function() {
                    updateChallenges();
                    currentChallenge = uncompletedChallenges()[0] ||
                      challenges[0];
                    if (checkForWin()) {
                      updateInfoText(
                          'Correct! You you completed all of the challenges, ' +
                          'but feel free to keep on going!', INFO_STATUS.correct
                      );
                    } else {
                      updateInfoText(
                          'Correct! You have a new challenge!',
                          INFO_STATUS.correct
                      );
                    }
                    routie('/' + currentChallenge.slug);
                  }
              );
            } else {
              updateInfoText(
                  'Incorrect answer, try again',
                  INFO_STATUS.incorrect
              );
              if (resp.test_errors && resp.test_errors.length > 0) {
                updateInfoText(
                    resp.test_errors[0] + ' - try again',
                    INFO_STATUS.incorrect
                );
              } else if (resp.rand_error) {
                updateInfoText(
                    'Test against random data failed - try again',
                    INFO_STATUS.incorrect
                );
              }
            }
          }
          term.resume();
        });
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
      onBeforeCommand: function(command) {
        $('.terminal').hide();
        $('#term-spinner').show();
      },
      onAfterCommand: function(command) {
        $('#term-spinner').hide();
        $('.terminal').show();
      },
      prompt: function(callback) {
        callback('(' + retCode + ')> ');
      },
      completion: function(string, callback) {
        const completions = Array.prototype.concat(
            TAB_COMPLETION, currentChallenge.completions || []);
        return completions;
      },
      doubleTab: function(completingString, completions, echoFn) {
        $('#completions').html(
            completions.join('<span style=\'color: grey;\'> | </span>')
        ).show().fadeOut(2500);
      },
      onClear: termClear,
    });
    term = $.terminal.active();
    updateRoutes();
  });
});
