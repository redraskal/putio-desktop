export interface go {
  "main": {
    "App": {
		CountDownloading():Promise<number>
		ListDownloads():Promise<Array<Download>>
		Log(arg1:string):Promise<void>
		Queue(arg1:string):Promise<void>
		ReportFile(arg1:string,arg2:string):Promise<void>
		ReportPath(arg1:string):Promise<void>
    },
  }

}

declare global {
	interface Window {
		go: go;
	}
}
