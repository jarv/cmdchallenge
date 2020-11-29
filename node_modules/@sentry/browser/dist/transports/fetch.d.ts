import { Event, Response, Session } from '@sentry/types';
import { BaseTransport } from './base';
/** `fetch` based transport */
export declare class FetchTransport extends BaseTransport {
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
//# sourceMappingURL=fetch.d.ts.map