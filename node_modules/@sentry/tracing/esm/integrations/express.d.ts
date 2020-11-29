import { Integration } from '@sentry/types';
declare type Method = 'all' | 'get' | 'post' | 'put' | 'delete' | 'patch' | 'options' | 'head' | 'checkout' | 'copy' | 'lock' | 'merge' | 'mkactivity' | 'mkcol' | 'move' | 'm-search' | 'notify' | 'purge' | 'report' | 'search' | 'subscribe' | 'trace' | 'unlock' | 'unsubscribe';
declare type Application = {
    [method in Method | 'use']: (...args: any) => any;
};
/**
 * Express integration
 *
 * Provides an request and error handler for Express framework
 * as well as tracing capabilities
 */
export declare class Express implements Integration {
    /**
     * @inheritDoc
     */
    static id: string;
    /**
     * @inheritDoc
     */
    name: string;
    /**
     * Express App instance
     */
    private readonly _app?;
    private readonly _methods?;
    /**
     * @inheritDoc
     */
    constructor(options?: {
        app?: Application;
        methods?: Method[];
    });
    /**
     * @inheritDoc
     */
    setupOnce(): void;
}
export {};
//# sourceMappingURL=express.d.ts.map