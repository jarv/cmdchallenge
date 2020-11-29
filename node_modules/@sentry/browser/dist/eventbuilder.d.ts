import { Event, EventHint, Options, Severity } from '@sentry/types';
/**
 * Builds and Event from a Exception
 * @hidden
 */
export declare function eventFromException(options: Options, exception: unknown, hint?: EventHint): PromiseLike<Event>;
/**
 * Builds and Event from a Message
 * @hidden
 */
export declare function eventFromMessage(options: Options, message: string, level?: Severity, hint?: EventHint): PromiseLike<Event>;
/**
 * @hidden
 */
export declare function eventFromUnknownInput(exception: unknown, syntheticException?: Error, options?: {
    rejection?: boolean;
    attachStacktrace?: boolean;
}): Event;
/**
 * @hidden
 */
export declare function eventFromString(input: string, syntheticException?: Error, options?: {
    attachStacktrace?: boolean;
}): Event;
//# sourceMappingURL=eventbuilder.d.ts.map