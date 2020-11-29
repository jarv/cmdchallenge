/*eslint-disable*/

const SENTRY_LOADER_RE = /sentry\.loader\.js$/;
const SENTRY_MODULE_RE = /sentry-webpack\.module\.js$/;

const mockCli = {
  releases: {
    new: jest.fn(() => Promise.resolve()),
    uploadSourceMaps: jest.fn(() => Promise.resolve()),
    finalize: jest.fn(() => Promise.resolve()),
    proposeVersion: jest.fn(() => Promise.resolve()),
    setCommits: jest.fn(() => Promise.resolve()),
  },
};

const SentryCliMock = jest.fn((configFile, options) => mockCli);
const SentryCli = jest.mock('@sentry/cli', () => SentryCliMock);
const SentryCliPlugin = require('../..');

afterEach(() => {
  jest.clearAllMocks();
});

const defaults = {
  debug: false,
  finalize: true,
  rewrite: true,
};

describe('constructor', () => {
  test('uses defaults without options', () => {
    const sentryCliPlugin = new SentryCliPlugin();

    expect(sentryCliPlugin.options).toEqual(defaults);
  });

  test('merges defaults with options', () => {
    const sentryCliPlugin = new SentryCliPlugin({
      foo: 42,
    });

    expect(sentryCliPlugin.options).toEqual(expect.objectContaining(defaults));
    expect(sentryCliPlugin.options.foo).toEqual(42);
  });

  test('uses declared options over defaults', () => {
    const sentryCliPlugin = new SentryCliPlugin({
      rewrite: false,
    });

    expect(sentryCliPlugin.options.rewrite).toEqual(false);
  });

  test('sanitizes array options `include` and `ignore`', () => {
    const sentryCliPlugin = new SentryCliPlugin({
      include: 'foo',
      ignore: 'bar',
    });
    expect(sentryCliPlugin.options.include).toEqual(['foo']);
    expect(sentryCliPlugin.options.ignore).toEqual(['bar']);
  });

  test('keeps array options `include` and `ignore`', () => {
    const sentryCliPlugin = new SentryCliPlugin({
      include: ['foo'],
      ignore: ['bar'],
    });
    expect(sentryCliPlugin.options.include).toEqual(['foo']);
    expect(sentryCliPlugin.options.ignore).toEqual(['bar']);
  });
});

describe('CLI configuration', () => {
  test('passes the configuration file to CLI', () => {
    const sentryCliPlugin = new SentryCliPlugin({
      configFile: 'some/sentry.properties',
    });

    expect(SentryCliMock).toHaveBeenCalledWith('some/sentry.properties', {
      silent: false,
    });
  });

  test('only creates a single CLI instance', () => {
    const sentryCliPlugin = new SentryCliPlugin({});
    sentryCliPlugin.apply({ hooks: { afterEmit: { tapAsync: jest.fn() } } });
    expect(SentryCliMock.mock.instances.length).toBe(1);
  });
});

describe('afterEmitHook', () => {
  let compiler;
  let compilation;
  let compilationDoneCallback;

  beforeEach(() => {
    compiler = {
      hooks: {
        afterEmit: {
          tapAsync: jest.fn((name, callback) =>
            callback(compilation, compilationDoneCallback)
          ),
        },
      },
    };

    compilation = { errors: [], hash: 'someHash' };
    compilationDoneCallback = jest.fn();
  });

  test('calls `hooks.afterEmit.tapAsync()`', () => {
    const sentryCliPlugin = new SentryCliPlugin();
    sentryCliPlugin.apply(compiler);

    expect(compiler.hooks.afterEmit.tapAsync).toHaveBeenCalledWith(
      'SentryCliPlugin',
      expect.any(Function)
    );
  });

  test('calls `compiler.plugin("after-emit")` legacy Webpack <= 3', () => {
    const sentryCliPlugin = new SentryCliPlugin();

    // Simulate Webpack <= 2
    compiler = { plugin: jest.fn() };
    sentryCliPlugin.apply(compiler);

    expect(compiler.plugin).toHaveBeenCalledWith(
      'after-emit',
      expect.any(Function)
    );
  });

  test('errors without `include` option', done => {
    const sentryCliPlugin = new SentryCliPlugin({ release: 42 });
    sentryCliPlugin.apply(compiler);

    setImmediate(() => {
      expect(compilationDoneCallback).toBeCalled();
      expect(compilation.errors).toEqual([
        new Error('Sentry CLI Plugin: `include` option is required'),
      ]);
      done();
    });
  });

  test('creates a release on Sentry', done => {
    expect.assertions(4);

    const sentryCliPlugin = new SentryCliPlugin({
      include: 'src',
      release: 42,
    });
    sentryCliPlugin.apply(compiler);

    setImmediate(() => {
      expect(mockCli.releases.new).toBeCalledWith('42');
      expect(mockCli.releases.uploadSourceMaps).toBeCalledWith(
        '42',
        expect.objectContaining({
          release: 42,
          include: ['src'],
        })
      );
      expect(mockCli.releases.finalize).toBeCalledWith('42');
      expect(compilationDoneCallback).toBeCalled();
      done();
    });
  });

  test('skips finalizing release if finalize:false', done => {
    expect.assertions(4);

    const sentryCliPlugin = new SentryCliPlugin({
      include: 'src',
      release: 42,
      finalize: false,
    });
    sentryCliPlugin.apply(compiler);

    setImmediate(() => {
      expect(mockCli.releases.new).toBeCalledWith('42');
      expect(mockCli.releases.uploadSourceMaps).toBeCalledWith(
        '42',
        expect.objectContaining({
          release: 42,
          include: ['src'],
        })
      );
      expect(mockCli.releases.finalize).not.toBeCalled();
      expect(compilationDoneCallback).toBeCalled();
      done();
    });
  });

  test('handles errors during releasing', done => {
    expect.assertions(2);
    mockCli.releases.new.mockImplementationOnce(() =>
      Promise.reject(new Error('Pickle Rick'))
    );

    const sentryCliPlugin = new SentryCliPlugin({
      include: 'src',
      release: 42,
    });
    sentryCliPlugin.apply(compiler);

    setImmediate(() => {
      expect(compilation.errors).toEqual([
        new Error('Sentry CLI Plugin: Pickle Rick'),
      ]);
      expect(compilationDoneCallback).toBeCalled();
      done();
    });
  });

  test('handles errors with errorHandler option', done => {
    expect.assertions(3);
    mockCli.releases.new.mockImplementationOnce(() =>
      Promise.reject(new Error('Pickle Rick'))
    );
    let e;

    const sentryCliPlugin = new SentryCliPlugin({
      include: 'src',
      release: 42,
      errorHandler: err => {
        e = err;
      },
    });
    sentryCliPlugin.apply(compiler);

    setImmediate(() => {
      expect(compilation.errors).toEqual([]);
      expect(e.message).toEqual('Pickle Rick');
      expect(compilationDoneCallback).toBeCalled();
      done();
    });
  });

  test('test setCommits with flat options', done => {
    const sentryCliPlugin = new SentryCliPlugin({
      include: 'src',
      release: '42',
      commit: '4d8656426ca13eab19581499da93408e30fdd9ef',
      previousCommit: 'b6b0e11e74fd55836d3299cef88531b2a34c2514',
      repo: 'group / repo',
      auto: false,
    });

    sentryCliPlugin.apply(compiler);

    setImmediate(() => {
      expect(mockCli.releases.setCommits).toBeCalledWith('42', {
        repo: 'group / repo',
        commit: '4d8656426ca13eab19581499da93408e30fdd9ef',
        previousCommit: 'b6b0e11e74fd55836d3299cef88531b2a34c2514',
        auto: false,
      });
      expect(compilationDoneCallback).toBeCalled();
      done();
    });
  });

  test('test setCommits with grouped options', done => {
    const sentryCliPlugin = new SentryCliPlugin({
      include: 'src',
      release: '42',
      setCommits: {
        commit: '4d8656426ca13eab19581499da93408e30fdd9ef',
        previousCommit: 'b6b0e11e74fd55836d3299cef88531b2a34c2514',
        repo: 'group / repo',
        auto: false,
      },
    });

    sentryCliPlugin.apply(compiler);

    setImmediate(() => {
      expect(mockCli.releases.setCommits).toBeCalledWith('42', {
        repo: 'group / repo',
        commit: '4d8656426ca13eab19581499da93408e30fdd9ef',
        previousCommit: 'b6b0e11e74fd55836d3299cef88531b2a34c2514',
        auto: false,
      });
      expect(compilationDoneCallback).toBeCalled();
      done();
    });
  });
});

describe('module rule overrides', () => {
  let compiler;
  let sentryCliPlugin;

  beforeEach(() => {
    sentryCliPlugin = new SentryCliPlugin({ release: '42', include: 'src' });
    compiler = {
      hooks: { afterEmit: { tapAsync: jest.fn() } },
      options: { module: {} },
    };
  });

  test('injects a `rule` for our mock module', () => {
    expect.assertions(1);

    compiler.options.module.rules = [];
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.module.rules[0]).toEqual({
      test: /sentry-webpack\.module\.js$/,
      use: [
        {
          loader: expect.stringMatching(SENTRY_LOADER_RE),
          options: { releasePromise: expect.any(Promise) },
        },
      ],
    });
  });

  test('injects a `loader` for our mock module', () => {
    expect.assertions(1);

    compiler.options.module.loaders = [];
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.module.loaders[0]).toEqual({
      test: /sentry-webpack\.module\.js$/,
      loader: expect.stringMatching(SENTRY_LOADER_RE),
      options: { releasePromise: expect.any(Promise) },
    });
  });

  test('defaults to `rules` when nothing is specified', () => {
    expect.assertions(1);
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.module.rules).toBeInstanceOf(Array);
  });

  test('creates the `module` option if missing', () => {
    expect.assertions(1);

    delete compiler.options.module;
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.module).not.toBeUndefined();
  });
});

describe('entry point overrides', () => {
  let compiler;
  let sentryCliPlugin;

  beforeEach(() => {
    sentryCliPlugin = new SentryCliPlugin({ release: '42', include: 'src' });
    compiler = {
      hooks: { afterEmit: { tapAsync: jest.fn() } },
      options: { module: { rules: [] } },
    };
  });

  test('creates an entry if none is specified', () => {
    expect.assertions(1);
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.entry).toMatch(SENTRY_MODULE_RE);
  });

  test('injects into a single entry', () => {
    expect.assertions(1);

    compiler.options.entry = './src/index.js';
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.entry).toEqual([
      expect.stringMatching(SENTRY_MODULE_RE),
      './src/index.js',
    ]);
  });

  test('injects into an array entry', () => {
    expect.assertions(1);

    compiler.options.entry = ['./src/preload.js', './src/index.js'];
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.entry).toEqual([
      expect.stringMatching(SENTRY_MODULE_RE),
      './src/preload.js',
      './src/index.js',
    ]);
  });

  test('injects into multiple entries', () => {
    expect.assertions(1);

    compiler.options.entry = {
      main: './src/index.js',
      admin: './src/admin.js',
    };
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.entry).toEqual({
      main: [expect.stringMatching(SENTRY_MODULE_RE), './src/index.js'],
      admin: [expect.stringMatching(SENTRY_MODULE_RE), './src/admin.js'],
    });
  });

  test('injects into multiple entries with array chunks', () => {
    expect.assertions(1);

    compiler.options.entry = {
      main: ['./src/index.js', './src/common.js'],
      admin: ['./src/admin.js', './src/common.js'],
    };
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.entry).toEqual({
      main: [
        expect.stringMatching(SENTRY_MODULE_RE),
        './src/index.js',
        './src/common.js',
      ],
      admin: [
        expect.stringMatching(SENTRY_MODULE_RE),
        './src/admin.js',
        './src/common.js',
      ],
    });
  });

  test('injects into entries specified by a function', () => {
    expect.assertions(1);

    compiler.options.entry = () => Promise.resolve('./src/index.js');
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.entry()).resolves.toEqual([
      expect.stringMatching(SENTRY_MODULE_RE),
      './src/index.js',
    ]);
  });

  test('injects into entries specified by entry descriptor with single import', () => {
    expect.assertions(1);

    compiler.options.entry = {
      main: {
        import: './src/index.js',
        out: '[name].bundle.js',
      },
    };
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.entry).toEqual({
      main: {
        import: [expect.stringMatching(SENTRY_MODULE_RE), './src/index.js'],
        out: '[name].bundle.js',
      },
    });
  });

  test('injects into entries specified by entry descriptor with multiple imports', () => {
    expect.assertions(1);

    compiler.options.entry = {
      main: {
        import: ['./src/index.js', './src/common.js'],
        out: '[name].bundle.js',
      },
    };
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.entry).toEqual({
      main: {
        import: [
          expect.stringMatching(SENTRY_MODULE_RE),
          './src/index.js',
          './src/common.js',
        ],
        out: '[name].bundle.js',
      },
    });
  });

  test('injects into entries specified by all possible methods at the same time', () => {
    expect.assertions(2);

    compiler.options.entry = {
      home: './home.js',
      about: ['./about.js'],
      contact: () => './contact.js',
      login: {
        import: './login.js',
      },
      logout: {
        import: ['./logout.js'],
      },
    };
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.entry).toEqual({
      home: [expect.stringMatching(SENTRY_MODULE_RE), './home.js'],
      about: [expect.stringMatching(SENTRY_MODULE_RE), './about.js'],
      contact: expect.any(Function),
      login: {
        import: [expect.stringMatching(SENTRY_MODULE_RE), './login.js'],
      },
      logout: {
        import: [expect.stringMatching(SENTRY_MODULE_RE), './logout.js'],
      },
    });
    expect(compiler.options.entry.contact()).resolves.toEqual([
      expect.stringMatching(SENTRY_MODULE_RE),
      './contact.js',
    ]);
  });

  test('filters entry points by name', () => {
    expect.assertions(1);

    compiler.options.entry = {
      main: './src/index.js',
      admin: './src/admin.js',
      login: {
        import: './src/login.js',
      },
    };
    sentryCliPlugin.options.entries = ['admin'];
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.entry).toEqual({
      main: './src/index.js',
      admin: [expect.stringMatching(SENTRY_MODULE_RE), './src/admin.js'],
      login: {
        import: './src/login.js',
      },
    });
  });

  test('filters entry points by RegExp', () => {
    expect.assertions(1);

    compiler.options.entry = {
      main: './src/index.js',
      admin: ['./src/admin.js', './src/common.js'],
      adminButBetter: {
        import: './src/admin.js',
      },
    };
    sentryCliPlugin.options.entries = /^ad/;
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.entry).toEqual({
      main: './src/index.js',
      admin: [
        expect.stringMatching(SENTRY_MODULE_RE),
        './src/admin.js',
        './src/common.js',
      ],
      adminButBetter: {
        import: [expect.stringMatching(SENTRY_MODULE_RE), './src/admin.js'],
      },
    });
  });

  test('filters entry points by function', () => {
    expect.assertions(1);

    compiler.options.entry = {
      main: ['./src/index.js', './src/common.js'],
      admin: './src/admin.js',
    };
    sentryCliPlugin.options.entries = key => key == 'admin';
    sentryCliPlugin.apply(compiler);

    expect(compiler.options.entry).toEqual({
      main: ['./src/index.js', './src/common.js'],
      admin: [expect.stringMatching(SENTRY_MODULE_RE), './src/admin.js'],
    });
  });

  test('throws for an invalid `entries` option', () => {
    compiler.options.entry = {
      main: './src/index.js',
      admin: './src/admin.js',
    };
    sentryCliPlugin.options.entries = 42;
    expect(() => sentryCliPlugin.apply(compiler)).toThrowError(/entries/);
  });
});
