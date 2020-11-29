import { Plugin, Compiler } from 'webpack';

export interface SentryCliPluginOptions {
  // general configuration for sentry-cli

  // Sentry instance
  url?: string;
  // Authentication token for API
  authToken?: string;
  // Organization slug
  org?: string;
  // Project slug
  project?: string;
  // VCS remote name
  vcsRemote?: string;

  /**
   * Unique name of a release, must be a string, should uniquely identify your release,
   * defaults to sentry-cli releases propose-version command which should always return the correct version
   * (requires access to git CLI and root directory to be a valid repository).
   */
  release?: string;

  /**
   * One or more paths that Sentry CLI should scan recursively for sources.
   * It will upload all .map files and match associated .js files.
   */
  include: string | string[];

  /**
   * A filter for entry points that should be processed.
   * By default, the release will be injected into all entry points.
   */
  entries?: string[] | RegExp | ((key: string) => boolean);

  /**
   * Path to a file containing list of files/directories to ignore.
   * Can point to .gitignore or anything with same format.
   */
  ignoreFile?: string;

  /**
   * One or more paths to ignore during upload. Overrides entries in ignoreFile file.
   * If neither ignoreFile or ignore are present, defaults to ['node_modules'].
   */
  ignore?: string | string[];

  /**
   * Path to Sentry CLI config properties, as described in https://docs.sentry.io/learn/cli/configuration/#properties-files.
   * By default, the config file is looked for upwards from the current path and defaults from ~/.sentryclirc are always loaded.
   */
  configFile?: string;

  /**
   * This sets the file extensions to be considered.
   * By default the following file extensions are processed: js, map, jsbundle and bundle.
   */
  ext?: string[];

  /**
   * This sets an URL prefix at the beginning of all files.
   * This defaults to `~/` but you might want to set this to the full URL.
   * This is also useful if your files are stored in a sub folder. eg: url-prefix `~/static/js`.
   */
  urlPrefix?: string;

  /**
   * This sets an URL suffix at the end of all files.
   * Useful for appending query parameters.
   */
  urlSuffix?: string;

  /**
   * This attempts sourcemap validation before upload when rewriting is not enabled.
   * It will spot a variety of issues with source maps and cancel the upload if any are found.
   * This is not the default as this can cause false positives.
   */
  validate?: boolean;

  /**
   * When paired with rewrite this will chop-off a prefix from uploaded files.
   * For instance you can use this to remove a path that is build machine specific.
   */
  stripPrefix?: string[];

  /**
   * When paired with rewrite this will add ~ to the stripPrefix array.
   */
  stripCommonPrefix?: boolean;

  /**
   * This prevents the automatic detection of sourcemap references.
   */
  sourceMapReference?: boolean;

  /**
   * Enables rewriting of matching sourcemaps so that indexed maps are flattened
   * and missing sources are inlined if possible., defaults to `true`.
   */
  rewrite?: boolean;

  /**
   * Determines whether processed release should be automatically finalized after artifacts upload.
   * Defaults to `true`.
   */
  finalize?: boolean;

  /**
   * Attempts a dry run (useful for dev environments).
   */
  dryRun?: boolean;

  /**
   * Print some useful debug information.
   */
  debug?: boolean;

  /**
   * If true, all logs are suppressed (useful for --json option).
   */
  silent?: boolean;

  /**
   * when Cli error occurs, plugin calls this function.
   * webpack compilation failure can be chosen by calling invokeErr callback or not.
   * defaults to `(err, invokeErr) => { invokeErr() }`
   */
  errorHandler?: (err: Error, invokeErr: () => void) => void;

  /**
   * Adds commits to sentry
   */
  setCommits?: {
    /**
     * The full repo name as defined in Sentry.
     * Required if auto option is not true.
     */
    repo?: string;

    /**
     * The current (last) commit in the release.
     * Required if auto option is not true.
     */
    commit?: string;

    /**
     * The commit before the beginning of this release (in other words, the last commit of the previous release).
     * If omitted, this will default to the last commit of the previous release in Sentry.
     * If there was no previous release, the last 10 commits will be used.
     */
    previousCommit?: string;

    /**
     * Automatically choose the associated commit (uses the current commit). Overrides other setCommit options.
     */
    auto?: boolean;
  };

  /**
   * Creates a new release deployment
   */
  deploy?: {
    /**
     * Environment for this release. Values that make sense here would be `production` or `staging`
     */
    env: string;

    /**
     * Unix timestamp when the deployment started
     */
    started?: number;

    /**
     * Unix timestamp when the deployment finished
     */
    finished?: number;

    /**
     * Deployment duration in seconds. This can be specified alternatively to `started` and `finished`
     */
    time?: number;

    /**
     * Human readable name for this deployment
     */
    name?: string;

    /**
     * URL that points to the deployment
     */
    url?: string;
  };
}

declare class SentryCliPlugin extends Plugin {
  constructor(options: SentryCliPluginOptions);

  apply(compiler: Compiler): void;
}

export default SentryCliPlugin;
