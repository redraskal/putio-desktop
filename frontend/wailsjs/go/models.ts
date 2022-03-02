/* Do not change, this code is generated from Golang structs */

export {};

export class Download {
    id: number;
    name: string;
    status: number;
    dl: number;
    total: number;

    static createFrom(source: any = {}) {
        return new Download(source);
    }

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.id = source["id"];
        this.name = source["name"];
        this.status = source["status"];
        this.dl = source["dl"];
        this.total = source["total"];
    }
}