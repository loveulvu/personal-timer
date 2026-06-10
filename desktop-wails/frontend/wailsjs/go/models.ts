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
	export class DailyTaskStat {
	    task_id: number;
	    title: string;
	    status: string;
	    estimated_minutes: number;
	    actual_seconds: number;
	    actual_minutes: number;
	
	    static createFrom(source: any = {}) {
	        return new DailyTaskStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.task_id = source["task_id"];
	        this.title = source["title"];
	        this.status = source["status"];
	        this.estimated_minutes = source["estimated_minutes"];
	        this.actual_seconds = source["actual_seconds"];
	        this.actual_minutes = source["actual_minutes"];
	    }
	}
	export class DailyStats {
	    date: string;
	    total_seconds: number;
	    total_minutes: number;
	    completed_count: number;
	    unfinished_count: number;
	    tasks: DailyTaskStat[];
	
	    static createFrom(source: any = {}) {
	        return new DailyStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.total_seconds = source["total_seconds"];
	        this.total_minutes = source["total_minutes"];
	        this.completed_count = source["completed_count"];
	        this.unfinished_count = source["unfinished_count"];
	        this.tasks = this.convertValues(source["tasks"], DailyTaskStat);
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
	
	export class GenerateSummaryResult {
	    summary_id: number;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new GenerateSummaryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.summary_id = source["summary_id"];
	        this.content = source["content"];
	    }
	}
	export class LLMTestResponse {
	    status: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new LLMTestResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.message = source["message"];
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
	export class Summary {
	    id: number;
	    summary_type: string;
	    start_date: string;
	    end_date: string;
	    content: string;
	    source_data?: any;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Summary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.summary_type = source["summary_type"];
	        this.start_date = source["start_date"];
	        this.end_date = source["end_date"];
	        this.content = source["content"];
	        this.source_data = source["source_data"];
	        this.created_at = source["created_at"];
	    }
	}
	
	export class WeeklyDayStat {
	    date: string;
	    total_seconds: number;
	    total_minutes: number;
	    completed_count: number;
	    unfinished_count: number;
	
	    static createFrom(source: any = {}) {
	        return new WeeklyDayStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.total_seconds = source["total_seconds"];
	        this.total_minutes = source["total_minutes"];
	        this.completed_count = source["completed_count"];
	        this.unfinished_count = source["unfinished_count"];
	    }
	}
	export class WeeklyProjectStat {
	    project_id: number;
	    project_name: string;
	    task_count: number;
	    completed_count: number;
	    total_seconds: number;
	    total_minutes: number;
	
	    static createFrom(source: any = {}) {
	        return new WeeklyProjectStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_id = source["project_id"];
	        this.project_name = source["project_name"];
	        this.task_count = source["task_count"];
	        this.completed_count = source["completed_count"];
	        this.total_seconds = source["total_seconds"];
	        this.total_minutes = source["total_minutes"];
	    }
	}
	export class WeeklyStats {
	    start_date: string;
	    end_date: string;
	    total_seconds: number;
	    total_minutes: number;
	    completed_count: number;
	    unfinished_count: number;
	    days: WeeklyDayStat[];
	    projects: WeeklyProjectStat[];
	
	    static createFrom(source: any = {}) {
	        return new WeeklyStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.start_date = source["start_date"];
	        this.end_date = source["end_date"];
	        this.total_seconds = source["total_seconds"];
	        this.total_minutes = source["total_minutes"];
	        this.completed_count = source["completed_count"];
	        this.unfinished_count = source["unfinished_count"];
	        this.days = this.convertValues(source["days"], WeeklyDayStat);
	        this.projects = this.convertValues(source["projects"], WeeklyProjectStat);
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

