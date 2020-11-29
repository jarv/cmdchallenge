import { Event, Response, Session } from '@sentry/types';
import { BaseTransport } from './base';
/** `XHR` based transport */
export declare class XHRTransport extends BaseTransport {
    /**
     * @inheritDoc
     */
    sendEvent(event: Event): PromiseLike<Response>;
    /**
     * @inheritDoc
     */
    sendSession(session: Session): PromiseLike<Response>;
    /**
     * @param sentryRequest Prepared SentryRequest to be delivered
     * @param originalPayload Original payload used to create SentryRequest
     */
    private _sendRequest;
}
//# sourceMappingURL=xhr.d.ts.map