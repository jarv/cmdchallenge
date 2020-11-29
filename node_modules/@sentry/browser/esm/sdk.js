import { getCurrentHub, initAndBind, Integrations as CoreIntegrations } from '@sentry/core';
import { getGlobalObject, SyncPromise } from '@sentry/utils';
import { BrowserClient } from './client';
import { wrap as internalWrap } from './helpers';
import { Breadcrumbs, GlobalHandlers, LinkedErrors, TryCatch, UserAgent } from './integrations';
export var defaultIntegrations = [
    new CoreIntegrations.InboundFilters(),
    new CoreIntegrations.FunctionToString(),
    new TryCatch(),
    new Breadcrumbs(),
    new GlobalHandlers(),
    new LinkedErrors(),
    new UserAgent(),
];
/**
 * The Sentry Browser SDK Client.
 *
 * To use this SDK, call the {@link init} function as early as possible when
 * loading the web page. To set context information or send manual events, use
 * the provided methods.
 *
 * @example
 *
 * ```
 *
 * import { init } from '@sentry/browser';
 *
 * init({
 *   dsn: '__DSN__',
 *   // ...
 * });
 * ```
 *
 * @example
 * ```
 *
 * import { configureScope } from '@sentry/browser';
 * configureScope((scope: Scope) => {
 *   scope.setExtra({ battery: 0.7 });
 *   scope.setTag({ user_mode: 'admin' });
 *   scope.setUser({ id: '4711' });
 * });
 * ```
 *
 * @example
 * ```
 *
 * import { addBreadcrumb } from '@sentry/browser';
 * addBreadcrumb({
 *   message: 'My Breadcrumb',
 *   // ...
 * });
 * ```
 *
 * @example
 *
 * ```
 *
 * import * as Sentry from '@sentry/browser';
 * Sentry.captureMessage('Hello, world!');
 * Sentry.captureException(new Error('Good bye'));
 * Sentry.captureEvent({
 *   message: 'Manual',
 *   stacktrace: [
 *     // ...
 *   ],
 * });
 * ```
 *
 * @see {@link BrowserOptions} for documentation on configuration options.
 */
export function init(options) {
    if (options === void 0) { options = {}; }
    if (options.defaultIntegrations === undefined) {
        options.defaultIntegrations = defaultIntegrations;
    }
    if (options.release === undefined) {
        var window_1 = getGlobalObject();
        // This supports the variable that sentry-webpack-plugin injects
        if (window_1.SENTRY_RELEASE && window_1.SENTRY_RELEASE.id) {
            options.release = window_1.SENTRY_RELEASE.id;
        }
    }
    if (options.autoSessionTracking === undefined) {
        options.autoSessionTracking = false;
    }
    initAndBind(BrowserClient, options);
    if (options.autoSessionTracking) {
        startSessionTracking();
    }
}
/**
 * Present the user with a report dialog.
 *
 * @param options Everything is optional, we try to fetch all info need from the global scope.
 */
export function showReportDialog(options) {
    if (options === void 0) { options = {}; }
    if (!options.eventId) {
        options.eventId = getCurrentHub().lastEventId();
    }
    var client = getCurrentHub().getClient();
    if (client) {
        client.showReportDialog(options);
    }
}
/**
 * This is the getter for lastEventId.
 *
 * @returns The last event id of a captured event.
 */
export function lastEventId() {
    return getCurrentHub().lastEventId();
}
/**
 * This function is here to be API compatible with the loader.
 * @hidden
 */
export function forceLoad() {
    // Noop
}
/**
 * This function is here to be API compatible with the loader.
 * @hidden
 */
export function onLoad(callback) {
    callback();
}
/**
 * A promise that resolves when all current events have been sent.
 * If you provide a timeout and the queue takes longer to drain the promise returns false.
 *
 * @param timeout Maximum time in ms the client should wait.
 */
export function flush(timeout) {
    var client = getCurrentHub().getClient();
    if (client) {
        return client.flush(timeout);
    }
    return SyncPromise.reject(false);
}
/**
 * A promise that resolves when all current events have been sent.
 * If you provide a timeout and the queue takes longer to drain the promise returns false.
 *
 * @param timeout Maximum time in ms the client should wait.
 */
export function close(timeout) {
    var client = getCurrentHub().getClient();
    if (client) {
        return client.close(timeout);
    }
    return SyncPromise.reject(false);
}
/**
 * Wrap code within a try/catch block so the SDK is able to capture errors.
 *
 * @param fn A function to wrap.
 *
 * @returns The result of wrapped function call.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function wrap(fn) {
    return internalWrap(fn)();
}
/**
 * Enable automatic Session Tracking for the initial page load.
 */
function startSessionTracking() {
    var window = getGlobalObject();
    var hub = getCurrentHub();
    /**
     * We should be using `Promise.all([windowLoaded, firstContentfulPaint])` here,
     * but, as always, it's not available in the IE10-11. Thanks IE.
     */
    var loadResolved = document.readyState === 'complete';
    var fcpResolved = false;
    var possiblyEndSession = function () {
        if (fcpResolved && loadResolved) {
            hub.endSession();
        }
    };
    var resolveWindowLoaded = function () {
        loadResolved = true;
        possiblyEndSession();
        window.removeEventListener('load', resolveWindowLoaded);
    };
    hub.startSession();
    if (!loadResolved) {
        // IE doesn't support `{ once: true }` for event listeners, so we have to manually
        // attach and then detach it once completed.
        window.addEventListener('load', resolveWindowLoaded);
    }
    try {
        var po = new PerformanceObserver(function (entryList, po) {
            entryList.getEntries().forEach(function (entry) {
                if (entry.name === 'first-contentful-paint' && entry.startTime < firstHiddenTime_1) {
                    po.disconnect();
                    fcpResolved = true;
                    possiblyEndSession();
                }
            });
        });
        // There's no need to even attach this listener if `PerformanceObserver` constructor will fail,
        // so we do it below here.
        var firstHiddenTime_1 = document.visibilityState === 'hidden' ? 0 : Infinity;
        document.addEventListener('visibilitychange', function (event) {
            firstHiddenTime_1 = Math.min(firstHiddenTime_1, event.timeStamp);
        }, { once: true });
        po.observe({
            type: 'paint',
            buffered: true,
        });
    }
    catch (e) {
        fcpResolved = true;
        possiblyEndSession();
    }
}
//# sourceMappingURL=sdk.js.map