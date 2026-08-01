export namespace backend {
	
	export class Model {
	    name: string;
	    description: string;
	    installed: boolean;
	    path?: string;
	    onnx_url?: string;
	    tokenizer_url?: string;
	    size_mb?: number;
	
	    static createFrom(source: any = {}) {
	        return new Model(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.installed = source["installed"];
	        this.path = source["path"];
	        this.onnx_url = source["onnx_url"];
	        this.tokenizer_url = source["tokenizer_url"];
	        this.size_mb = source["size_mb"];
	    }
	}
	export class Project {
	    name: string;
	    path: string;
	    star: boolean;
	    lastModified: string;
	    status?: number;
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
	export class Response {
	    statusCode: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new Response(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statusCode = source["statusCode"];
	        this.message = source["message"];
	    }
	}

}

