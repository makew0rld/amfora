# package `rr`, aka `RestartReader`

This package exists just to hold the `RestartReader` type. It wraps `io.ReadCloser` and implements it. It holds the data from every `Read` in a `[]byte` buffer, and allows you to call `.Restart()`, causing subsequent `Read` calls to start from the beginning again.

See [#140](https://github.com/makeworld-the-better-one/amfora/issues/140) for why this was needed.

Other projects are encouraged to copy this code if it's useful to them, and this package may move out of Amfora if I end up using it in multiple projects.

## License

If you prefer, you can consider the code in this package, and this package only, to be licensed under the MIT license instead.

<details>
<summary>Click to see MIT license terms</summary>

```
Copyright (c) 2020 makeworld

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
</details>
