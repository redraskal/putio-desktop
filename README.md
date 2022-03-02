# putio-desktop
A desktop client for [put.io](https://put.io) supporting multi-connection downloads & resuming downloads (soon). **The client is not yet complete.**

![putio-desktop-dev_tnr2lgKl0o](https://user-images.githubusercontent.com/6241454/156285218-d5df17f1-138b-448d-a288-d896b42b6c61.png)

## Install
- [Download the latest release](https://github.com/redraskal/putio-desktop/releases).

## Roadmap for v1
- [ ] Support resuming downloads
- [ ] Support pausing downloads from frontend
- [ ] Support VLC playlist download prompt
- [ ] Cleanup frontend code

## Live Development
Install wails v2 cli with the following command:
```
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```
See the [wails documentation](https://wails.io/docs/gettingstarted/installation) for more details.

To run in live development mode, run `wails dev` in the project directory.

**At the moment, you must re-run the dev command to see frontend changes.**

## Building
For a production build, use `wails build`.

## Thanks
* Huge thanks to [@leaanthony](https://github.com/leaanthony) and contributors over at https://github.com/wailsapp/wails for the platform this project runs on
