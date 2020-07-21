/* eslint strict: ["error", "global"] */

'use strict';
const HOSTNAME = window.location.hostname.split('.');
const BASEURL = HOSTNAME.filter((i) => i != "oops").join(".")
const OOPS_IMG = '<img src="img/emojis/1F92D.png" alt="" />';
const CMD_IMG = '<img src="img/cmdchallenge-round.png" alt="" />';

const BASEURLS = {
  CMD: '//' + BASEURL,
  OOPS: '//oops.' + BASEURL,
}
const SITES = {
  CMD: "cmdchallenge",
  OOPS: "oops"
}
const FLAVOR = HOSTNAME[0].match(/oops/) !== null ? SITES.OOPS : SITES.CMD
const SITE_LINKS = {
  CMD: '<a href="//' + BASEURLS.CMD + '">' + CMD_IMG + '</a>',
  OOPS: '<a href="//' + BASEURLS.OOPS + '">' + OOPS_IMG + '</a>',
}

const CMD_URL = window.location.hostname.match(/local/) ? 'https://testing.cmdchallenge.com/r' : '/r';
const TAB_COMPLETION = FLAVOR === SITES.OOPS ? ['echo', 'read'] : ['find', 'echo', 'awk', 'sed', 'perl', 'wc', 'grep', 'cat', 'sort', 'cut', 'ls', 'tac', 'jq', 'paste', 'tr', 'rm', 'tail', 'comm', 'egrep'];

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

  const stepGen = function* (steps) {
    while (true) yield* steps;
  };

  const errorEmoji=stepGen(['1F63F.png']);
  const incorrectEmoji=stepGen(
      ['E282.png', '1F645-200D-2640-FE0F.png', '1F645-200D-2642-FE0F.png', '1F940.png']
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
      ['1F478.png', '1F482.png', '1F9DD.png', '1F9DD-200D-2640-FE0F.png', '1F680.png']
  );

  const correctEmojiOops=stepGen(
      ['1F600.png', '1F604.png', '1F970.png', '1F60D.png', '1F929.png']
  );


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
      switch (FLAVOR) {
        case SITES.OOPS:
          $('.title .won').html('🎉 Congrats, you completed the challenge! 🎉 Try <a href="' + BASEURLS.CMD + '">even more challenges!</a>').show()
          break;
        case SITES.CMD:
          $('.title .won').html('🎉 Congrats, you completed the challenge! 🎉').show()
          break;
      }
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
      routes['/' + slug] = function() {
        currentChallenge = c;
        updateChallengeDesc();
        updateChallenges();
        checkForWin();
      };
    });
    routes[''] = function() {
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
            slug + '\'><img class=\'badge\' src=\'img/emojis/' +
            c.openemoji + '.png\' alt=\'' +
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

    displaySolution()
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
          switch (FLAVOR) {
            case SITES.OOPS:
              callback(resp.filter((o) => (o.tags || []).includes('oops')))
              break;
            case SITES.CMD:
              callback(resp.filter((o) => !(o.tags)));
              break;
          }
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
      'img': currentChallenge.img || 'cmd',
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

  const updateInfoText = function(msg, infoStatus) {
    $('#info-box .text').html(msg);
    switch (infoStatus) {
      case INFO_STATUS.correct:
        const index = challenges.indexOf(currentChallenge);
        let emojiFname;

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
    $('#challenge-desc .img-container').html('<img src=\'img/emojis/' +
        currentChallenge.openemoji +
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
    $.ajax({
      dataType: 'json',
      url: '/s/solutions/' + currentChallenge.slug + '.json',
      success: function(resp) {
        $('#solutions').html('').show();
        resp.cmds.forEach(function(cmd) {
          $('#solutions').append(escapeHtml(cmd) + '\n');
        });
        hljs.highlightBlock(document.getElementById('solutions'));
        $('#solutions-wrapper .last-updated').html(
            'Solutions updated ' + dateDelta(resp.ts) + ' ago'
        );
      },
      error: function() {
        $('#solutions').html('').hide();
        $('#solutions-wrapper .last-updated').html(
            'No solutions for this challenge yet'
        );
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

  const processResp = function(resp) {
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
  };

  // main

  // Setup for different site types
  switch (FLAVOR) {
    case SITES.OOPS:
      $('#header-img').html(OOPS_IMG);
      $('#header-text').html('Oops I deleted my bin/ dir :(');
      $('#links ul').prepend(SITE_LINKS.CMD);
      break;
    case SITES.CMD:
      $('#links ul').prepend(SITE_LINKS.OOPS);
      break;
  }

  $('#header-img').show();
  $('#header-text').show();


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
        if (/tail\s+-[Ff]/.test(command)) {
          updateInfoText(
              '<code>tail -f</code> will wait for additional data ' +
                'to be appended to the file, try removing the -f option',
              INFO_STATUS.incorrect
          );
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
        ).show().fadeOut(5000);
      },
      onClear: termClear,
    });
    term = $.terminal.active();
    updateRoutes();
  });
});
