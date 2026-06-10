export namespace api {
	
	export class ConfigStatus {
	    database: string;
	    llm_configured: boolean;
	    llm_base_url_configured: boolean;
	    llm_model_configured: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.database = source["database"];
	        this.llm_configured = source["llm_configured"];
	        this.llm_base_url_configured = source["llm_base_url_configured"];
	        this.llm_model_configured = source["llm_model_configured"];
	        this.error = source["error"];
	    }
	}
	export class CreateDailyTaskRequest {
	    project_id?: number;
	    task_date: string;
	    title: string;
	    estimated_minutes: number;
	
	    static createFrom(source: any = {}) {
	        return new CreateDailyTaskRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_id = source["project_id"];
	        this.task_date = source["task_date"];
	        this.title = source["title"];
	        this.estimated_minutes = source["estimated_minutes"];
	    }
	}
	export class CreateResponse {
	    id: number;
	
	    static createFrom(source: any = {}) {
	        return new CreateResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	    }
	}
	export class DailyTask {
	    id: number;
	    project_id?: number;
	    task_date: string;
	    title: string;
	    estimated_minutes: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new DailyTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.project_id = source["project_id"];
	        this.task_date = source["task_date"];
	        this.title = source["title"];
	        this.estimated_minutes = source["estimated_minutes"];
	        this.status = source["status"];
	    }
	}
	export class Project {
	    id: number;
	    name: string;
	    description: string;
	    is_fixed: boolean;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.is_fixed = source["is_fixed"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProjectInput {
	    name: string;
	    description: string;
	    is_fixed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProjectInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.is_fixed = source["is_fixed"];
	    }
	}
	export class VersionInfo {
	    name: string;
	    version: string;
	    mode: string;
	
	    static createFrom(source: any = {}) {
	        return new VersionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.mode = source["mode"];
	    }
	}
	export class StartupStatus {
	    connected: boolean;
	    version?: VersionInfo;
	    config?: ConfigStatus;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new StartupStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.version = this.convertValues(source["version"], VersionInfo);
	        this.config = this.convertValues(source["config"], ConfigStatus);
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

