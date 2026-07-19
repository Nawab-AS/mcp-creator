export namespace backend {
	
	export class Model {
	    name: string;
	    description: string;
	    installed: boolean;
	    path?: string;
	
	    static createFrom(source: any = {}) {
	        return new Model(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.installed = source["installed"];
	        this.path = source["path"];
	    }
	}
	export class Project {
	    name: string;
	    path: string;
	    star: boolean;
	    lastModified: string;
	    status: number;
	    modelUsed: string;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.star = source["star"];
	        this.lastModified = source["lastModified"];
	        this.status = source["status"];
	        this.modelUsed = source["modelUsed"];
	    }
	}

}

