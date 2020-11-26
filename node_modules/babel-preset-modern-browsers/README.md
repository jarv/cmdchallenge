<h3 align="center">
  babel-preset-modern-browsers
</h3>

<p align="center">
  Babel presets for modern browsers
</p>

<p align="center">
  <a href="https://npmjs.org/package/babel-preset-modern-browsers"><img src="https://img.shields.io/npm/v/babel-preset-modern-browsers.svg?style=flat-square"></a>
</p>

- [Installation](#installation)
- [Usage](#usage)
- [Presets](#presets)
- [Compatibility Table](#compatibility-table)
- [Release Dates](#release-dates)

This preset covers syntax of `es2015`, `es2016`, `es2017`, `es2018`, `es2019` and `es2020`.

More info in the compatibility table below

# babel 7

Since v12, this package requires `@babel/core@7.0.0`. If you use babel 6, you can still use the version "11.0.1" of this package. If you want to migrate, you can read the [announcement](https://babeljs.io/blog/2018/08/27/7.0.0) and the [official migration guide](https://babeljs.io/docs/en/v7-migration).

## Alternatives

- [@babel/preset-env](https://www.npmjs.com/package/@babel/preset-env), especially `targets.esmodules`

If you don't need preset-env, using this package will only install a few dependencies.

## Modern browsers

![Edge 83][edge-83] ![Firefox 78][firefox-78] ![Chrome 80][chrome-80] ![Opera 67][opera-67] ![Safari 13.1][safari-13.1]

## Installation

```sh
npm install --save-dev babel-preset-modern-browsers @babel/core
```

## Usage

Add the following line to your `.babelrc` file:

```js
{
  "presets": ["modern-browsers"]
}
```

### Options

- `loose`: Enable “loose” transformations for any plugins in this preset that allow them (Disabled by default).
- `modules` - Enable transformation of ES6 module syntax to another module type (Enabled by default to "commonjs"). Can be false to not transform modules, or "commonjs"
- `shippedProposals` - Enable features in stages but already available in browsers (Enabled by default)

```js
{
  presets: [['modern-browsers', { loose: true }]];
}
```

```js
{
  presets: [[require('babel-preset-modern-browsers'), { loose: true }]];
}
```

## Browserlist

```
Edge >= 83, Firefox >= 78, FirefoxAndroid  >= 78, Chrome >= 80, ChromeAndroid >= 80, Opera >= 67, OperaMobile >= 67, Safari >= 13.1, iOS >= 13.4
```

## Compatibility Table

| Feature                                                                                                                        | Edge                | Firefox                   | Chrome                  | Opera                 | Safari                      |
| ------------------------------------------------------------------------------------------------------------------------------ | ------------------- | ------------------------- | ----------------------- | --------------------- | --------------------------- |
| <h3>Shipped Proposals</h3>                                                                                                     |                     |                           |                         |                       |                             |
| [Numeric Separators](http://kangax.github.io/compat-table/es2016plus/#test-numeric_separators)                                 | ![Edge 79][edge-79] | ![Firefox 70][firefox-70] | ![Chrome 75][chrome-75] | ![Opera 62][opera-62] | ![Safari 13][safari-13]     |
| ↳ [syntax-numeric-separator](https://www.npmjs.com/package/@babel/plugin-syntax-numeric-separator)                             |                     |                           |                         |                       |                             |
| <h3>ES2020</h3>                                                                                                                |                     |                           |                         |                       |                             |
| [Optional chaining (`?.`)](<http://kangax.github.io/compat-table/es2016plus/#test-optional_chaining_operator_(?.)>)            | ![Edge 80][edge-80] | ![Firefox 74][firefox-74] | ![Chrome 80][chrome-80] | ![Opera 67][opera-67] | ![Safari 13.1][safari-13.1] |
| [Nullish Coalescing operator (`??`)](<http://kangax.github.io/compat-table/es2016plus/#test-nullish_coalescing_operator_(??)>) | ![Edge 80][edge-80] | ![Firefox 72][firefox-72] | ![Chrome 80][chrome-80] | ![Opera 67][opera-67] | ![Safari 13.1][safari-13.1] |
| <h3>ES2019</h3>                                                                                                                |                     |                           |                         |                       |                             |
| [Optional catch binding](http://kangax.github.io/compat-table/es2016plus/#test-optional_catch_binding)                         | ![Edge 79][edge-79] | ![Firefox 58][firefox-58] | ![Chrome 66][chrome-66] | ![Opera 53][opera-53] | ![Safari 11.1][safari-11.1] |
| [JSON strings](http://kangax.github.io/compat-table/es2016plus/#test-JSON_superset)                                            | ![Edge 79][edge-79] | ![Firefox 62][firefox-62] | ![Chrome 66][chrome-66] | ![Opera 53][opera-53] | ![Safari 12][safari-12]     |
| <h3>ES2018</h3>                                                                                                                |                     |                           |                         |                       |                             |
| [Object Rest/Spread Properties](https://kangax.github.io/compat-table/es2016plus/#test-object_rest/spread_properties)          | ![Edge 79][edge-79] | ![Firefox 55][firefox-55] | ![Chrome 60][chrome-60] | ![Opera 47][opera-47] | ![Safari 11.1][safari-11.1] |
| [RegExp Unicode Property Escapes](https://kangax.github.io/compat-table/es2016plus/#test-RegExp_Unicode_Property_Escapes)      | ![Edge 79][edge-79] | ![Firefox 78][firefox-78] | ![Chrome 64][chrome-64] | ![Opera 51][opera-51] | ![Safari 11.1][safari-11.1] |
| [Asynchronous Iterators](https://kangax.github.io/compat-table/es2016plus/#test-Asynchronous_Iterators)                        | ![Edge 79][edge-79] | ![Firefox 57][firefox-57] | ![Chrome 63][chrome-63] | ![Opera 50][opera-50] | ![Safari 12][safari-12]     |
| <h3>ES2017</h3>                                                                                                                |                     |                           |                         |                       |                             |
| [trailing commas in function](http://kangax.github.io/compat-table/es2016plus/#test-trailing_commas_in_function_syntax)        | ![Edge 14][edge-14] | ![Firefox 52][firefox-52] | ![Chrome 58][chrome-58] | ![Opera 45][opera-45] | ![Safari 10][safari-10]     |
| [async function](http://kangax.github.io/compat-table/es2016plus/#test-async_functions)                                        | ![Edge 15][edge-15] | ![Firefox 52][firefox-52] | ![Chrome 55][chrome-55] | ![Opera 42][opera-42] | ![Safari 10.1][safari-10.1] |
| <h3>ES2016</h3>                                                                                                                |                     |                           |                         |                       |                             |
| [exponentiation operator](<http://kangax.github.io/compat-table/es2016plus/#test-exponentiation_(**)_operator>)                | ![Edge 14][edge-14] | ![Firefox 52][firefox-52] | ![Chrome 52][chrome-52] | ![Opera 39][opera-39] | ![Safari 10][safari-10]     |
| <h3>ES2015</h3>                                                                                                                | ![Edge 79][edge-79] | ![Firefox 53][firefox-53] | ![Chrome 52][chrome-52] | ![Opera 39][opera-39] | ![Safari 10][safari-10]     |
| <h4>Syntax</h4>                                                                                                                |                     |                           |                         |                       |                             |
| [default parameters](https://kangax.github.io/compat-table/es6/#test-default_function_parameters)                              | ![Edge 14][edge-14] | ![Firefox 53][firefox-53] | ![Chrome 49][chrome-49] | ![Opera 36][opera-36] | ![Safari 10][safari-10]     |
| [rest parameters](https://kangax.github.io/compat-table/es6/#test-rest_parameters)                                             | ![Edge 12][edge-12] | ![Firefox 43][firefox-43] | ![Chrome 47][chrome-47] | ![Opera 34][opera-34] | ![Safari 10][safari-10]     |
| [spread](https://kangax.github.io/compat-table/es6/#test-spread)                                                               | ![Edge 13][edge-13] | ![Firefox 36][firefox-36] | ![Chrome 46][chrome-46] | ![Opera 33][opera-33] | ![Safari 10][safari-10]     |
| [computed properties](https://kangax.github.io/compat-table/es6/#test-object_literal_extensions_computed_properties)           | ![Edge 12][edge-12] | ![Firefox 34][firefox-34] | ![Chrome 44][chrome-44] | ![Opera 31][opera-31] | ![Safari 7.1][safari-7.1]   |
| [shorthand properties](https://kangax.github.io/compat-table/es6/#test-object_literal_extensions_shorthand_properties)         | ![Edge 12][edge-12] | ![Firefox 33][firefox-33] | ![Chrome 43][chrome-43] | ![Opera 30][opera-30] | ![Safari 9][safari-9]       |
| [`for...of`](https://kangax.github.io/compat-table/es6/#test-for..of_loops)                                                    | ![Edge 14][edge-14] | ![Firefox 53][firefox-53] | ![Chrome 51][chrome-51] | ![Opera 38][opera-38] | ![Safari 10][safari-10]     |
| [template string](https://kangax.github.io/compat-table/es6/#test-template_strings)                                            | ![Edge 13][edge-13] | ![Firefox 34][firefox-34] | ![Chrome 41][chrome-41] | ![Opera 28][opera-28] | ![Safari 9][safari-9]       |
| [Regexp sticky](https://kangax.github.io/compat-table/es6/#test-RegExp_y_and_u_flags_y_flag)                                   | ![Edge 13][edge-13] | ![Firefox 31][firefox-31] | ![Chrome 49][chrome-49] | ![Opera 36][opera-36] | ![Safari 10][safari-10]     |
| [Regexp unicode](https://kangax.github.io/compat-table/es6/#test-RegExp_y_and_u_flags_u_flag)                                  | ![Edge 12][edge-12] | ![Firefox 46][firefox-46] | ![Chrome 51][chrome-51] | ![Opera 38][opera-38] | ![Safari 10][safari-10]     |
| [destructuring](https://kangax.github.io/compat-table/es6/)                                                                    | ![Edge 15][edge-15] | ![Firefox 53][firefox-53] | ![Chrome 52][chrome-52] | ![Opera 39][opera-39] | ![Safari 10][safari-10]     |
| [Unicode Strings](https://kangax.github.io/compat-table/es6/#test-Unicode_code_point_escapes_in_strings)                       | ![Edge 12][edge-12] | ![Firefox 45][firefox-45] | ![Chrome 44][chrome-44] | ![Opera 31][opera-31] | ![Safari 9][safari-9]       |
| [Octal/Binary Numbers](https://kangax.github.io/compat-table/es6/#test-octal_and_binary_literals)                              | ![Edge 12][edge-12] | ![Firefox 36][firefox-36] | ![Chrome 41][chrome-41] | ![Opera 28][opera-28] | ![Safari 9][safari-9]       |
| <h4>Bindings</h4>                                                                                                              |                     |                           |                         |                       |                             |
| [`const`](https://kangax.github.io/compat-table/es6/#test-const)                                                               | ![Edge 14][edge-14] | ![Firefox 51][firefox-51] | ![Chrome 49][chrome-49] | ![Opera 36][opera-36] | ![Safari 10][safari-10]     |
| [`let`](https://kangax.github.io/compat-table/es6/#test-let)                                                                   | ![Edge 14][edge-14] | ![Firefox 51][firefox-51] | ![Chrome 49][chrome-49] | ![Opera 36][opera-36] | ![Safari 10][safari-10]     |
| [`block-level function declaration`](https://kangax.github.io/compat-table/es6/#test-block-level_function_declaration)         | ![Edge 11][edge-11] | ![Firefox 46][firefox-46] | ![Chrome 41][chrome-41] | ![Opera 28][opera-28] | ![Safari 10][safari-10]     |
| <h4>Functions</h4>                                                                                                             |                     |                           |                         |                       |                             |
| [arrow functions](https://kangax.github.io/compat-table/es6/#test-arrow_functions)                                             | ![Edge 13][edge-13] | ![Firefox 45][firefox-45] | ![Chrome 49][chrome-49] | ![Opera 36][opera-36] | ![Safari 10][safari-10]     |
| [classes](https://kangax.github.io/compat-table/es6/#test-arrow_functions)                                                     | ![Edge 13][edge-13] | ![Firefox 45][firefox-45] | ![Chrome 49][chrome-49] | ![Opera 36][opera-36] | ![Safari 10][safari-10]     |
| [super](https://kangax.github.io/compat-table/es6/#test-super)                                                                 | ![Edge 13][edge-13] | ![Firefox 45][firefox-45] | ![Chrome 49][chrome-49] | ![Opera 36][opera-36] | ![Safari 10][safari-10]     |
| [generators](https://kangax.github.io/compat-table/es6/#test-generators)                                                       | ![Edge 13][edge-13] | ![Firefox 53][firefox-53] | ![Chrome 51][chrome-51] | ![Opera 38][opera-38] | ![Safari 10][safari-10]     |
| <h4>Built-ins</h4>                                                                                                             |                     |                           |                         |                       |                             |
| [typeof Symbol](https://kangax.github.io/compat-table/es6/#test-Symbol_typeof_support)                                         | ![Edge 12][edge-12] | ![Firefox 36][firefox-36] | ![Chrome 38][chrome-38] | ![Opera 25][opera-25] | ![Safari 9][safari-9]       |
| <h4>Built-in extensions</h4>                                                                                                   |                     |                           |                         |                       |                             |
| [function name](https://kangax.github.io/compat-table/es6/#test-function_name_property)                                        | ![Edge 79][edge-79] | ![Firefox 53][firefox-53] | ![Chrome 52][chrome-52] | ![Opera 39][opera-39] | ![Safari 10][safari-10]     |

## Partially Shipped Proposals (Not included)

| Feature                                                                                                                                                                            | Edge                | Firefox                       | Chrome                  | Opera                 | Safari                      |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------- | ----------------------------- | ----------------------- | --------------------- | --------------------------- |
| [Static](http://kangax.github.io/compat-table/esnext/#test-static_class_fields) & [Instance](http://kangax.github.io/compat-table/esnext/#test-instance_class_fields) Class Fields | ![Edge 79][edge-79] | ![Firefox None][firefox-none] | ![Chrome 74][chrome-74] | ![Opera 61][opera-61] | ![Safari None][safari-none] |
| ↳ [proposal-class-properties](https://www.npmjs.com/package/@babel/plugin-proposal-class-properties)                                                                               |                     |                               |                         |                       |                             |
| [Private Class Methods](http://kangax.github.io/compat-table/esnext/#test-private_class_methods)                                                                                   | ![Edge 84][edge-84] | ![Firefox None][firefox-none] | ![Chrome 84][chrome-84] | ![Opera 71][opera-71] | ![Safari None][safari-none] |
| ↳ [proposal-private-methods](https://www.npmjs.com/package/@babel/plugin-proposal-private-methods)                                                                                 |                     |                               |                         |                       |                             |

## Release Dates

- Firefox: https://wiki.mozilla.org/RapidRelease/Calendar
- Chrome: https://www.chromium.org/developers/calendar ([Version History](https://en.wikipedia.org/wiki/Google_Chrome_version_history))
- Safari: https://developer.apple.com/safari/ ([Version History](https://en.wikipedia.org/wiki/Safari_version_history))
- Edge: https://developer.microsoft.com/en-us/microsoft-edge/platform/changelog/ ([Version History](https://en.wikipedia.org/wiki/Microsoft_Edge#Release_history))

## Thanks

- Inspired by [https://github.com/askmatey/babel-preset-modern](https://github.com/askmatey/babel-preset-modern)

[edge-11]: https://img.shields.io/badge/Edge-11-green.svg?style=flat-square
[edge-12]: https://img.shields.io/badge/Edge-12-green.svg?style=flat-square
[edge-13]: https://img.shields.io/badge/Edge-13-green.svg?style=flat-square
[edge-14]: https://img.shields.io/badge/Edge-14-green.svg?style=flat-square
[edge-15]: https://img.shields.io/badge/Edge-15-green.svg?style=flat-square
[edge-16]: https://img.shields.io/badge/Edge-16-green.svg?style=flat-square
[edge-17]: https://img.shields.io/badge/Edge-17-green.svg?style=flat-square
[edge-18]: https://img.shields.io/badge/Edge-18-green.svg?style=flat-square
[edge-79]: https://img.shields.io/badge/Edge-79-green.svg?style=flat-square
[edge-80]: https://img.shields.io/badge/Edge-80-green.svg?style=flat-square
[edge-83]: https://img.shields.io/badge/Edge-83-green.svg?style=flat-square
[edge-84]: https://img.shields.io/badge/Edge-84-green.svg?style=flat-square
[firefox-31]: https://img.shields.io/badge/Firefox-31-green.svg?style=flat-square
[firefox-33]: https://img.shields.io/badge/Firefox-33-green.svg?style=flat-square
[firefox-34]: https://img.shields.io/badge/Firefox-34-green.svg?style=flat-square
[firefox-36]: https://img.shields.io/badge/Firefox-36-green.svg?style=flat-square
[firefox-43]: https://img.shields.io/badge/Firefox-43-green.svg?style=flat-square
[firefox-45]: https://img.shields.io/badge/Firefox-45-green.svg?style=flat-square
[firefox-46]: https://img.shields.io/badge/Firefox-46-green.svg?style=flat-square
[firefox-47]: https://img.shields.io/badge/Firefox-47-green.svg?style=flat-square
[firefox-48]: https://img.shields.io/badge/Firefox-48-green.svg?style=flat-square
[firefox-49]: https://img.shields.io/badge/Firefox-49-green.svg?style=flat-square
[firefox-50]: https://img.shields.io/badge/Firefox-50-green.svg?style=flat-square
[firefox-51]: https://img.shields.io/badge/Firefox-51-green.svg?style=flat-square
[firefox-52]: https://img.shields.io/badge/Firefox-52-green.svg?style=flat-square
[firefox-53]: https://img.shields.io/badge/Firefox-53-green.svg?style=flat-square
[firefox-54]: https://img.shields.io/badge/Firefox-54-green.svg?style=flat-square
[firefox-55]: https://img.shields.io/badge/Firefox-55-green.svg?style=flat-square
[firefox-56]: https://img.shields.io/badge/Firefox-56-green.svg?style=flat-square
[firefox-56]: https://img.shields.io/badge/Firefox-56-green.svg?style=flat-square
[firefox-57]: https://img.shields.io/badge/Firefox-57-green.svg?style=flat-square
[firefox-58]: https://img.shields.io/badge/Firefox-58-green.svg?style=flat-square
[firefox-59]: https://img.shields.io/badge/Firefox-59-green.svg?style=flat-square
[firefox-60]: https://img.shields.io/badge/Firefox-60-green.svg?style=flat-square
[firefox-61]: https://img.shields.io/badge/Firefox-61-green.svg?style=flat-square
[firefox-62]: https://img.shields.io/badge/Firefox-62-green.svg?style=flat-square
[firefox-63]: https://img.shields.io/badge/Firefox-63-green.svg?style=flat-square
[firefox-64]: https://img.shields.io/badge/Firefox-64-green.svg?style=flat-square
[firefox-65]: https://img.shields.io/badge/Firefox-65-green.svg?style=flat-square
[firefox-70]: https://img.shields.io/badge/Firefox-70-green.svg?style=flat-square
[firefox-72]: https://img.shields.io/badge/Firefox-72-green.svg?style=flat-square
[firefox-74]: https://img.shields.io/badge/Firefox-74-green.svg?style=flat-square
[firefox-78]: https://img.shields.io/badge/Firefox-78-green.svg?style=flat-square
[firefox-none]: https://img.shields.io/badge/Firefox-None-red.svg?style=flat-square
[chrome-38]: https://img.shields.io/badge/Chrome-38-green.svg?style=flat-square
[chrome-39]: https://img.shields.io/badge/Chrome-39-green.svg?style=flat-square
[chrome-41]: https://img.shields.io/badge/Chrome-41-green.svg?style=flat-square
[chrome-43]: https://img.shields.io/badge/Chrome-43-green.svg?style=flat-square
[chrome-44]: https://img.shields.io/badge/Chrome-44-green.svg?style=flat-square
[chrome-46]: https://img.shields.io/badge/Chrome-46-green.svg?style=flat-square
[chrome-47]: https://img.shields.io/badge/Chrome-47-green.svg?style=flat-square
[chrome-49]: https://img.shields.io/badge/Chrome-49-green.svg?style=flat-square
[chrome-51]: https://img.shields.io/badge/Chrome-51-green.svg?style=flat-square
[chrome-52]: https://img.shields.io/badge/Chrome-52-green.svg?style=flat-square
[chrome-53]: https://img.shields.io/badge/Chrome-53-green.svg?style=flat-square
[chrome-54]: https://img.shields.io/badge/Chrome-54-green.svg?style=flat-square
[chrome-55]: https://img.shields.io/badge/Chrome-55-green.svg?style=flat-square
[chrome-56]: https://img.shields.io/badge/Chrome-56-green.svg?style=flat-square
[chrome-57]: https://img.shields.io/badge/Chrome-57-green.svg?style=flat-square
[chrome-58]: https://img.shields.io/badge/Chrome-58-green.svg?style=flat-square
[chrome-59]: https://img.shields.io/badge/Chrome-59-green.svg?style=flat-square
[chrome-60]: https://img.shields.io/badge/Chrome-60-green.svg?style=flat-square
[chrome-61]: https://img.shields.io/badge/Chrome-61-green.svg?style=flat-square
[chrome-62]: https://img.shields.io/badge/Chrome-62-green.svg?style=flat-square
[chrome-63]: https://img.shields.io/badge/Chrome-63-green.svg?style=flat-square
[chrome-64]: https://img.shields.io/badge/Chrome-64-green.svg?style=flat-square
[chrome-65]: https://img.shields.io/badge/Chrome-65-green.svg?style=flat-square
[chrome-66]: https://img.shields.io/badge/Chrome-66-green.svg?style=flat-square
[chrome-67]: https://img.shields.io/badge/Chrome-67-green.svg?style=flat-square
[chrome-68]: https://img.shields.io/badge/Chrome-68-green.svg?style=flat-square
[chrome-69]: https://img.shields.io/badge/Chrome-69-green.svg?style=flat-square
[chrome-70]: https://img.shields.io/badge/Chrome-70-green.svg?style=flat-square
[chrome-71]: https://img.shields.io/badge/Chrome-71-green.svg?style=flat-square
[chrome-72]: https://img.shields.io/badge/Chrome-72-green.svg?style=flat-square
[chrome-74]: https://img.shields.io/badge/Chrome-74-green.svg?style=flat-square
[chrome-75]: https://img.shields.io/badge/Chrome-75-green.svg?style=flat-square
[chrome-80]: https://img.shields.io/badge/Chrome-80-green.svg?style=flat-square
[chrome-84]: https://img.shields.io/badge/Chrome-84-green.svg?style=flat-square
[chrome-canary]: https://img.shields.io/badge/Chrome%20Canary-72-red.svg?style=flat-square
[opera-25]: https://img.shields.io/badge/Opera-25-green.svg?style=flat-square
[opera-26]: https://img.shields.io/badge/Opera-26-green.svg?style=flat-square
[opera-28]: https://img.shields.io/badge/Opera-28-green.svg?style=flat-square
[opera-30]: https://img.shields.io/badge/Opera-30-green.svg?style=flat-square
[opera-31]: https://img.shields.io/badge/Opera-31-green.svg?style=flat-square
[opera-33]: https://img.shields.io/badge/Opera-33-green.svg?style=flat-square
[opera-34]: https://img.shields.io/badge/Opera-34-green.svg?style=flat-square
[opera-36]: https://img.shields.io/badge/Opera-36-green.svg?style=flat-square
[opera-38]: https://img.shields.io/badge/Opera-38-green.svg?style=flat-square
[opera-39]: https://img.shields.io/badge/Opera-39-green.svg?style=flat-square
[opera-42]: https://img.shields.io/badge/Opera-42-green.svg?style=flat-square
[opera-45]: https://img.shields.io/badge/Opera-45-green.svg?style=flat-square
[opera-47]: https://img.shields.io/badge/Opera-47-green.svg?style=flat-square
[opera-50]: https://img.shields.io/badge/Opera-50-green.svg?style=flat-square
[opera-51]: https://img.shields.io/badge/Opera-51-green.svg?style=flat-square
[opera-52]: https://img.shields.io/badge/Opera-52-green.svg?style=flat-square
[opera-53]: https://img.shields.io/badge/Opera-53-green.svg?style=flat-square
[opera-61]: https://img.shields.io/badge/Opera-61-green.svg?style=flat-square
[opera-62]: https://img.shields.io/badge/Opera-62-green.svg?style=flat-square
[opera-67]: https://img.shields.io/badge/Opera-67-green.svg?style=flat-square
[opera-71]: https://img.shields.io/badge/Opera-71-green.svg?style=flat-square
[safari-7.1]: https://img.shields.io/badge/Safari-7.1-green.svg?style=flat-square
[safari-9]: https://img.shields.io/badge/Safari-9-green.svg?style=flat-square
[safari-10]: https://img.shields.io/badge/Safari-10-green.svg?style=flat-square
[safari-10.1]: https://img.shields.io/badge/Safari-10.1-green.svg?style=flat-square
[safari-11]: https://img.shields.io/badge/Safari-11-green.svg?style=flat-square
[safari-11.1]: https://img.shields.io/badge/Safari-11.1-green.svg?style=flat-square
[safari-12]: https://img.shields.io/badge/Safari-12-green.svg?style=flat-square
[safari-13]: https://img.shields.io/badge/Safari-13-green.svg?style=flat-square
[safari-13.1]: https://img.shields.io/badge/Safari-13.1-green.svg?style=flat-square
[safari-none]: https://img.shields.io/badge/Safari-None-red.svg?style=flat-square
