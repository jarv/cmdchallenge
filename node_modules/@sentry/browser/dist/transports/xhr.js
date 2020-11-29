Object.defineProperty(exports, "__esModule", { value: true });
var tslib_1 = require("tslib");
var core_1 = require("@sentry/core");
var utils_1 = require("@sentry/utils");
var base_1 = require("./base");
/** `XHR` based transport */
var XHRTransport = /** @class */ (function (_super) {
    tslib_1.__extends(XHRTransport, _super);
    function XHRTransport() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    /**
     * @inheritDoc
     */
    XHRTransport.prototype.sendEvent = function (event) {
        return this._sendRequest(core_1.eventToSentryRequest(event, this._api), event);
    };
    /**
     * @inheritDoc
     */
    XHRTransport.prototype.sendSession = function (session) {
        return this._sendRequest(core_1.sessionToSentryRequest(session, this._api), session);
    };
    /**
     * @param sentryRequest Prepared SentryRequest to be delivered
     * @param originalPayload Original payload used to create SentryRequest
     */
    XHRTransport.prototype._sendRequest = function (sentryRequest, originalPayload) {
        var _this = this;
        if (this._isRateLimited(sentryRequest.type)) {
            return Promise.reject({
                event: originalPayload,
                type: sentryRequest.type,
                reason: "Transport locked till " + this._disabledUntil(sentryRequest.type) + " due to too many requests.",
                status: 429,
            });
        }
        return this._buffer.add(new utils_1.SyncPromise(function (resolve, reject) {
            var request = new XMLHttpRequest();
            request.onreadystatechange = function () {
                if (request.readyState === 4) {
                    var headers = {
                        'x-sentry-rate-limits': request.getResponseHeader('X-Sentry-Rate-Limits'),
                        'retry-after': request.getResponseHeader('Retry-After'),
                    };
                    _this._handleResponse({ requestType: sentryRequest.type, response: request, headers: headers, resolve: resolve, reject: reject });
                }
            };
            request.open('POST', sentryRequest.url);
            for (var header in _this.options.headers) {
                if (_this.options.headers.hasOwnProperty(header)) {
                    request.setRequestHeader(header, _this.options.headers[header]);
                }
            }
            request.send(sentryRequest.body);
        }));
    };
    return XHRTransport;
}(base_1.BaseTransport));
exports.XHRTransport = XHRTransport;
//# sourceMappingURL=xhr.js.map