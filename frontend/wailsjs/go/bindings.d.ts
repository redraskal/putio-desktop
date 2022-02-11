export interface go {
  "main": {
    "App": {
		Log(arg1:string):Promise<void>
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
