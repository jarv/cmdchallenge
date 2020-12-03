"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = loader;

var _loaderUtils = require("loader-utils");

var _schemaUtils = require("schema-utils");

var _options = _interopRequireDefault(require("./options.json"));

var _utils = require("./utils");

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

/*
  MIT License http://www.opensource.org/licenses/mit-license.php
  Author Tobias Koppers @sokra
*/
function loader() {
  const options = (0, _loaderUtils.getOptions)(this);
  (0, _schemaUtils.validate)(_options.default, options, {
    name: "Expose Loader",
    baseDataPath: "options"
  });
  const callback = this.async();
  let exposes;

  try {
    exposes = (0, _utils.getExposes)(options.exposes);
  } catch (error) {
    callback(error);
    return;
  }
  /*
   * Workaround until module.libIdent() in webpack/webpack handles this correctly.
   *
   * Fixes:
   * - https://github.com/webpack-contrib/expose-loader/issues/55
   * - https://github.com/webpack-contrib/expose-loader/issues/49
   */


  this._module.userRequest = (0, _utils.getNewUserRequest)(this._module.userRequest);
  /*
   * Adding side effects
   *
   * Fixes:
   * - https://github.com/webpack-contrib/expose-loader/issues/120
   */

  if (this._module.factoryMeta) {
    this._module.factoryMeta.sideEffectFree = false;
  } // Change the request from an /abolute/path.js to a relative ./path.js.
  // This prevents [chunkhash] values from changing when running webpack builds in different directories.


  const newRequest = (0, _utils.contextify)(this.context, (0, _loaderUtils.getRemainingRequest)(this));
  const stringifiedNewRequest = (0, _loaderUtils.stringifyRequest)(this, `-!${newRequest}`);
  let code = `var ___EXPOSE_LOADER_IMPORT___ = require(${stringifiedNewRequest});\n`;
  code += `var ___EXPOSE_LOADER_GET_GLOBAL_THIS___ = require(${(0, _loaderUtils.stringifyRequest)(this, require.resolve("./runtime/getGlobalThis.js"))});\n`;
  code += `var ___EXPOSE_LOADER_GLOBAL_THIS___ = ___EXPOSE_LOADER_GET_GLOBAL_THIS___;\n`;

  for (const expose of exposes) {
    const {
      globalName,
      moduleLocalName,
      override
    } = expose;
    const globalNameInterpolated = globalName.map(item => (0, _loaderUtils.interpolateName)(this, item, {}));

    if (typeof moduleLocalName !== "undefined") {
      code += `var ___EXPOSE_LOADER_IMPORT_MODULE_LOCAL_NAME___ = ___EXPOSE_LOADER_IMPORT___.${moduleLocalName}\n`;
    }

    let propertyString = "___EXPOSE_LOADER_GLOBAL_THIS___";

    for (let i = 0; i < globalName.length; i++) {
      if (i > 0) {
        code += `if (typeof ${propertyString} === 'undefined') ${propertyString} = {};\n`;
      }

      propertyString += `[${JSON.stringify(globalNameInterpolated[i])}]`;
    }

    if (!override) {
      code += `if (typeof ${propertyString} === 'undefined') `;
    }

    code += typeof moduleLocalName !== "undefined" ? `${propertyString} = ___EXPOSE_LOADER_IMPORT_MODULE_LOCAL_NAME___;\n` : `${propertyString} = ___EXPOSE_LOADER_IMPORT___;\n`;

    if (!override) {
      if (this.mode === "development") {
        code += `else throw new Error('[exposes-loader] The "${globalName.join(".")}" value exists in the global scope, it may not be safe to overwrite it, use the "override" option')\n`;
      }
    }
  }

  code += `module.exports = ___EXPOSE_LOADER_IMPORT___;\n`;
  callback(null, code);
}