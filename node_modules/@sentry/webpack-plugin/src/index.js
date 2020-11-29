const SentryCli = require('@sentry/cli');
const path = require('path');
const util = require('util');

const SENTRY_LOADER = path.resolve(__dirname, 'sentry.loader.js');
const SENTRY_MODULE = path.resolve(__dirname, 'sentry-webpack.module.js');

/**
 * Helper function that ensures an object key is defined. This mutates target!
 *
 * @param {object} target The target object
 * @param {string} key The object key
 * @param {function} factory A function that creates the new element
 * @returns {any} The existing or created element.
 */
function ensure(target, key, factory) {
  // eslint-disable-next-line no-param-reassign
  target[key] = typeof target[key] !== 'undefined' ? target[key] : factory();
  return target[key];
}

/** Deep copy of a given input */
function sillyClone(input) {
  try {
    return JSON.parse(JSON.stringify(input));
  } catch (oO) {
    return undefined;
  }
}

/** Diffs two arrays */
function diffArray(prev, next) {
  // eslint-disable-next-line no-param-reassign
  prev = Array.isArray(prev) ? prev : [prev];
  // eslint-disable-next-line no-param-reassign
  next = Array.isArray(next) ? next : [next];

  return {
    removed: prev.filter(x => !next.includes(x)),
    added: next.filter(x => !prev.includes(x)),
  };
}

/** Extracts loader's name independently of Webpack's version */
function getLoaderName(entry) {
  return (
    entry.loader ||
    (entry.use && entry.use[0] && entry.use[0].loader) ||
    '<unknown loader>'
  );
}

/**
 * Ensures that the passed value is in an array or an array itself.
 *
 * @param {any} value Either an array or a value that should be wrapped
 * @returns {array} The array
 */
function toArray(value) {
  return !value || Array.isArray(value) ? value : [value];
}

/** Backwards compatible version of `compiler.plugin.afterEmit.tapAsync()`. */
function attachAfterEmitHook(compiler, callback) {
  if (compiler.hooks && compiler.hooks.afterEmit) {
    compiler.hooks.afterEmit.tapAsync('SentryCliPlugin', callback);
  } else {
    compiler.plugin('after-emit', callback);
  }
}

class SentryCliPlugin {
  constructor(options = {}) {
    const defaults = {
      debug: false,
      finalize: true,
      rewrite: true,
    };

    this.options = Object.assign({}, defaults, options);

    if (options.include) this.options.include = toArray(options.include);
    if (options.ignore) this.options.ignore = toArray(options.ignore);

    this.cli = this.getSentryCli();
    this.release = this.getReleasePromise();
  }

  /**
   * Pretty-prints debug information
   *
   * @param {string} label Label to be printed as a prefix for the data
   * @param {any} data Input to be pretty-printed
   */
  outputDebug(label, data) {
    if (this.isSilent()) {
      return;
    }
    if (data !== undefined) {
      // eslint-disable-next-line no-console
      console.log(
        `[Sentry Webpack Plugin] ${label} ${util.inspect(
          data,
          false,
          null,
          true
        )}`
      );
    } else {
      // eslint-disable-next-line no-console
      console.log(`[Sentry Webpack Plugin] ${label}`);
    }
  }

  /** Returns whether this plugin should emit any data to stdout. */
  isSilent() {
    return this.options.silent === true;
  }

  /** Returns whether this plugin is in dryRun mode. */
  isDryRun() {
    return this.options.dryRun === true;
  }

  /** Creates a new Sentry CLI instance. */
  getSentryCli() {
    const cli = new SentryCli(this.options.configFile, {
      silent: this.isSilent(),
      org: this.options.org,
      project: this.options.project,
      authToken: this.options.authToken,
      url: this.options.url,
      vcsRemote: this.options.vcsRemote,
    });

    if (this.isDryRun()) {
      this.outputDebug('DRY Run Mode');

      return {
        releases: {
          proposeVersion: () =>
            cli.releases.proposeVersion().then(version => {
              this.outputDebug('Proposed version:\n', version);
              return version;
            }),
          new: release => {
            this.outputDebug('Creating new release:\n', release);
            return Promise.resolve(release);
          },
          uploadSourceMaps: (release, config) => {
            this.outputDebug('Calling upload-sourcemaps with:\n', config);
            return Promise.resolve(release, config);
          },
          finalize: release => {
            this.outputDebug('Finalizing release:\n', release);
            return Promise.resolve(release);
          },
          setCommits: (release, config) => {
            this.outputDebug('Calling set-commits with:\n', config);
            return Promise.resolve(release, config);
          },
          newDeploy: (release, config) => {
            this.outputDebug('Calling deploy with:\n', config);
            return Promise.resolve(release, config);
          },
        },
      };
    }

    return cli;
  }

  /**
   * Returns a Promise that will solve to the configured release.
   *
   * If no release is specified, it uses Sentry CLI to propose a version.
   * The release string is always trimmed.
   * Returns undefined if proposeVersion failed.
   */
  getReleasePromise() {
    return (this.options.release
      ? Promise.resolve(this.options.release)
      : this.cli.releases.proposeVersion()
    )
      .then(version => `${version}`.trim())
      .catch(() => undefined);
  }

  /** Checks if the given named entry point should be handled. */
  shouldInjectEntry(key) {
    const { entries } = this.options;
    if (entries == null) {
      return true;
    }

    if (typeof entries === 'function') {
      return entries(key);
    }

    if (entries instanceof RegExp) {
      return entries.test(key);
    }

    if (Array.isArray(entries)) {
      return entries.includes(key);
    }

    throw new Error(
      'Invalid `entries` option: Must be an array, RegExp or function'
    );
  }

  /** Injects the release string into the given entry point. */
  injectEntry(entry, sentryModule) {
    if (!entry) {
      return sentryModule;
    }

    /**
     * in:
     *   entry: 'index.js'
     * out:
     *   entry: ['sentry-webpack.module.js', 'index.js']
     */
    if (typeof entry === 'string') {
      return [sentryModule, entry];
    }

    /**
     * in:
     *   entry: ['index.js', 'header.js', 'footer.js']
     * out:
     *   entry: ['sentry-webpack.module.js', 'index.js', 'header.js', 'footer.js']
     */
    if (Array.isArray(entry)) {
      return [sentryModule].concat(entry);
    }

    /**
     * in:
     *   entry: () => 'index.js'
     *   entry: () => ['index.js']
     * out:
     *   entry: ['sentry-webpack.module.js', 'index.js']
     *   entry: ['sentry-webpack.module.js', 'index.js']
     */
    if (typeof entry === 'function') {
      return () =>
        Promise.resolve(entry()).then(resolvedEntry =>
          this.injectEntry(resolvedEntry, sentryModule)
        );
    }

    /**
     * in:
     *   entry: {
     *     home: './home.js',
     *     about: ['./about.js'],
     *     contact: () => './contact.js',
     *     login: {
     *       import: './login.js',
     *     },
     *     logout: {
     *       import: ['./logout.js']
     *     }
     *   }
     * out:
     *   entry: {
     *     home: ['sentry-webpack.module.js', './home.js'],
     *     about: ['sentry-webpack.module.js', './about.js'],
     *     contact: ['sentry-webpack.module.js', './contact.js'],
     *     login: {
     *       import: ['sentry-webpack.module.js', './login.js']
     *     },
     *     logout: {
     *       import: ['sentry-webpack.module.js', './logout.js']
     *     }
     *   }
     */
    const modifiedEntry = { ...entry };
    Object.keys(modifiedEntry)
      .filter(key => this.shouldInjectEntry(key))
      .forEach(key => {
        if (entry[key] && entry[key].import) {
          modifiedEntry[key].import = this.injectEntry(
            entry[key].import,
            sentryModule
          );
        } else {
          modifiedEntry[key] = this.injectEntry(entry[key], sentryModule);
        }
      });
    return modifiedEntry;
  }

  /** Webpack 2: Adds a new loader for the release module. */
  injectLoader(loaders) {
    const loader = {
      test: /sentry-webpack\.module\.js$/,
      loader: SENTRY_LOADER,
      options: {
        releasePromise: this.release,
      },
    };

    return (loaders || []).concat([loader]);
  }

  /** Webpack 3+: Injects a new rule for the release module. */
  injectRule(rules) {
    const rule = {
      test: /sentry-webpack\.module\.js$/,
      use: [
        {
          loader: SENTRY_LOADER,
          options: {
            releasePromise: this.release,
          },
        },
      ],
    };

    return (rules || []).concat([rule]);
  }

  /** Injects the release entry points and rules into the given options. */
  injectRelease(compilerOptions) {
    const options = compilerOptions;
    options.entry = this.injectEntry(options.entry, SENTRY_MODULE);
    if (options.module.loaders) {
      // Handle old `options.module.loaders` syntax
      options.module.loaders = this.injectLoader(options.module.loaders);
    } else {
      options.module.rules = this.injectRule(options.module.rules);
    }
  }

  /** injectRelease with printable debug info */
  injectReleaseWithDebug(compilerOptions) {
    const input = {
      loaders: sillyClone(
        compilerOptions.module.loaders || compilerOptions.module.rules
      ).map(getLoaderName),
      entry: sillyClone(compilerOptions.entry),
    };

    this.injectRelease(compilerOptions);

    const output = {
      loaders: sillyClone(
        compilerOptions.module.loaders || compilerOptions.module.rules
      ).map(getLoaderName),
      entry: sillyClone(compilerOptions.entry),
    };

    const loaders = diffArray(input.loaders, output.loaders);
    const entry = diffArray(input.entry, output.entry);

    this.outputDebug('DEBUG: Injecting release code');
    this.outputDebug('DEBUG: Loaders:\n', output.loaders);
    this.outputDebug('DEBUG: Added loaders:\n', loaders.added);
    this.outputDebug('DEBUG: Removed loaders:\n', loaders.removed);
    this.outputDebug('DEBUG: Entry:\n', output.entry);
    this.outputDebug('DEBUG: Added entry:\n', entry.added);
    this.outputDebug('DEBUG: Removed entry:\n', entry.removed);
  }

  /** Creates and finalizes a release on Sentry. */
  finalizeRelease(compilation) {
    const {
      include,
      errorHandler = (_, invokeErr) => {
        invokeErr();
      },
    } = this.options;

    let release;
    return this.release
      .then(proposedVersion => {
        release = proposedVersion;

        if (!include) {
          throw new Error(`\`include\` option is required`);
        }

        if (!release) {
          throw new Error(
            `Unable to determine version. Make sure to include \`release\` option or use the environment that supports auto-detection https://docs.sentry.io/cli/releases/#creating-releases`
          );
        }

        return this.cli.releases.new(release);
      })
      .then(() => this.cli.releases.uploadSourceMaps(release, this.options))
      .then(() => {
        const { commit, previousCommit, repo, auto } =
          this.options.setCommits || this.options;

        if (auto || (repo && commit)) {
          return this.cli.releases.setCommits(release, {
            commit,
            previousCommit,
            repo,
            auto,
          });
        }

        return undefined;
      })
      .then(() => {
        if (this.options.finalize) {
          return this.cli.releases.finalize(release);
        }
        return undefined;
      })
      .then(() => {
        const { env, started, finished, time, name, url } =
          this.options.deploy || {};

        if (env) {
          return this.cli.releases.newDeploy(release, {
            env,
            started,
            finished,
            time,
            name,
            url,
          });
        }
        return undefined;
      })
      .catch(err => {
        errorHandler(
          err,
          () =>
            compilation.errors.push(
              new Error(`Sentry CLI Plugin: ${err.message}`)
            ),
          compilation
        );
      });
  }

  /** Webpack lifecycle hook to update compiler options. */
  apply(compiler) {
    const compilerOptions = compiler.options || {};
    ensure(compilerOptions, 'module', Object);

    if (this.options.debug) {
      this.injectReleaseWithDebug(compilerOptions);
    } else {
      this.injectRelease(compilerOptions);
    }

    attachAfterEmitHook(compiler, (compilation, cb) => {
      this.finalizeRelease(compilation).then(() => cb());
    });
  }
}

module.exports.default = SentryCliPlugin;
