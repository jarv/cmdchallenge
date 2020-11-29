import { BrowserTracing } from './browser';
import { addExtensionMethods } from './hubextensions';
import * as TracingIntegrations from './integrations';
declare const Integrations: {
    BrowserTracing: typeof BrowserTracing;
    Express: typeof TracingIntegrations.Express;
};
export { Integrations };
export { Span } from './span';
export { Transaction } from './transaction';
export { SpanStatus } from './spanstatus';
export { addExtensionMethods };
export { extractTraceparentData, getActiveTransaction, hasTracingEnabled, stripUrlQueryAndFragment } from './utils';
//# sourceMappingURL=index.d.ts.map