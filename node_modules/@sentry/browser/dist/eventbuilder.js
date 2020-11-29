Object.defineProperty(exports, "__esModule", { value: true });
var tslib_1 = require("tslib");
var types_1 = require("@sentry/types");
var utils_1 = require("@sentry/utils");
var parsers_1 = require("./parsers");
var tracekit_1 = require("./tracekit");
/**
 * Builds and Event from a Exception
 * @hidden
 */
function eventFromException(options, exception, hint) {
    var syntheticException = (hint && hint.syntheticException) || undefined;
    var event = eventFromUnknownInput(exception, syntheticException, {
        attachStacktrace: options.attachStacktrace,
    });
    utils_1.addExceptionMechanism(event, {
        handled: true,
        type: 'generic',
    });
    event.level = types_1.Severity.Error;
    if (hint && hint.event_id) {
        event.event_id = hint.event_id;
    }
    return utils_1.SyncPromise.resolve(event);
}
exports.eventFromException = eventFromException;
/**
 * Builds and Event from a Message
 * @hidden
 */
function eventFromMessage(options, message, level, hint) {
    if (level === void 0) { level = types_1.Severity.Info; }
    var syntheticException = (hint && hint.syntheticException) || undefined;
    var event = eventFromString(message, syntheticException, {
        attachStacktrace: options.attachStacktrace,
    });
    event.level = level;
    if (hint && hint.event_id) {
        event.event_id = hint.event_id;
    }
    return utils_1.SyncPromise.resolve(event);
}
exports.eventFromMessage = eventFromMessage;
/**
 * @hidden
 */
function eventFromUnknownInput(exception, syntheticException, options) {
    if (options === void 0) { options = {}; }
    var event;
    if (utils_1.isErrorEvent(exception) && exception.error) {
        // If it is an ErrorEvent with `error` property, extract it to get actual Error
        var errorEvent = exception;
        // eslint-disable-next-line no-param-reassign
        exception = errorEvent.error;
        event = parsers_1.eventFromStacktrace(tracekit_1.computeStackTrace(exception));
        return event;
    }
    if (utils_1.isDOMError(exception) || utils_1.isDOMException(exception)) {
        // If it is a DOMError or DOMException (which are legacy APIs, but still supported in some browsers)
        // then we just extract the name, code, and message, as they don't provide anything else
        // https://developer.mozilla.org/en-US/docs/Web/API/DOMError
        // https://developer.mozilla.org/en-US/docs/Web/API/DOMException
        var domException = exception;
        var name_1 = domException.name || (utils_1.isDOMError(domException) ? 'DOMError' : 'DOMException');
        var message = domException.message ? name_1 + ": " + domException.message : name_1;
        event = eventFromString(message, syntheticException, options);
        utils_1.addExceptionTypeValue(event, message);
        if ('code' in domException) {
            event.tags = tslib_1.__assign(tslib_1.__assign({}, event.tags), { 'DOMException.code': "" + domException.code });
        }
        return event;
    }
    if (utils_1.isError(exception)) {
        // we have a real Error object, do nothing
        event = parsers_1.eventFromStacktrace(tracekit_1.computeStackTrace(exception));
        return event;
    }
    if (utils_1.isPlainObject(exception) || utils_1.isEvent(exception)) {
        // If it is plain Object or Event, serialize it manually and extract options
        // This will allow us to group events based on top-level keys
        // which is much better than creating new group when any key/value change
        var objectException = exception;
        event = parsers_1.eventFromPlainObject(objectException, syntheticException, options.rejection);
        utils_1.addExceptionMechanism(event, {
            synthetic: true,
        });
        return event;
    }
    // If none of previous checks were valid, then it means that it's not:
    // - an instance of DOMError
    // - an instance of DOMException
    // - an instance of Event
    // - an instance of Error
    // - a valid ErrorEvent (one with an error property)
    // - a plain Object
    //
    // So bail out and capture it as a simple message:
    event = eventFromString(exception, syntheticException, options);
    utils_1.addExceptionTypeValue(event, "" + exception, undefined);
    utils_1.addExceptionMechanism(event, {
        synthetic: true,
    });
    return event;
}
exports.eventFromUnknownInput = eventFromUnknownInput;
/**
 * @hidden
 */
function eventFromString(input, syntheticException, options) {
    if (options === void 0) { options = {}; }
    var event = {
        message: input,
    };
    if (options.attachStacktrace && syntheticException) {
        var stacktrace = tracekit_1.computeStackTrace(syntheticException);
        var frames_1 = parsers_1.prepareFramesForEvent(stacktrace.stack);
        event.stacktrace = {
            frames: frames_1,
        };
    }
    return event;
}
exports.eventFromString = eventFromString;
//# sourceMappingURL=eventbuilder.js.map