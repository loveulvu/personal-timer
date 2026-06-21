export namespace api {
	
	export class DailyTask {
	    id: number;
	    project_id?: number;
	    task_date: string;
	    title: string;
	    estimated_minutes: number;
	    status: string;
	    finish_note?: string;
	    finish_description?: string;
	    // Go type: time
	    completed_at?: any;
	    actual_seconds_override?: number;
	    actual_seconds: number;
	    // Go type: time
	    current_session_started_at?: any;
	
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
	        this.finish_note = source["finish_note"];
	        this.finish_description = source["finish_description"];
	        this.completed_at = this.convertValues(source["completed_at"], null);
	        this.actual_seconds_override = source["actual_seconds_override"];
	        this.actual_seconds = source["actual_seconds"];
	        this.current_session_started_at = this.convertValues(source["current_session_started_at"], null);
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
	export class AcceptActionItemResult {
	    created: boolean;
	    already_exists: boolean;
	    task?: DailyTask;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new AcceptActionItemResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.created = source["created"];
	        this.already_exists = source["already_exists"];
	        this.task = this.convertValues(source["task"], DailyTask);
	        this.message = source["message"];
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
	
	
	export class EstimatePreviewRequest {
	    project_id: number;
	    title: string;
	    estimated_minutes: number;
	
	    static createFrom(source: any = {}) {
	        return new EstimatePreviewRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_id = source["project_id"];
	        this.title = source["title"];
	        this.estimated_minutes = source["estimated_minutes"];
	    }
	}
	export class EstimatePreviewResponse {
	    project_id: number;
	    input_estimated_minutes: number;
	    sample_count: number;
	    avg_estimated_minutes: number;
	    avg_actual_minutes: number;
	    overrun_rate: number;
	    risk_level: string;
	    suggested_minutes: number;
	    split_recommended: boolean;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new EstimatePreviewResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_id = source["project_id"];
	        this.input_estimated_minutes = source["input_estimated_minutes"];
	        this.sample_count = source["sample_count"];
	        this.avg_estimated_minutes = source["avg_estimated_minutes"];
	        this.avg_actual_minutes = source["avg_actual_minutes"];
	        this.overrun_rate = source["overrun_rate"];
	        this.risk_level = source["risk_level"];
	        this.suggested_minutes = source["suggested_minutes"];
	        this.split_recommended = source["split_recommended"];
	        this.reason = source["reason"];
	    }
	}
	export class FeedbackRequest {
	    target_type: string;
	    target_id: number;
	    target_index?: number;
	    feedback_value: string;
	    feedback_note: string;
	
	    static createFrom(source: any = {}) {
	        return new FeedbackRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.target_type = source["target_type"];
	        this.target_id = source["target_id"];
	        this.target_index = source["target_index"];
	        this.feedback_value = source["feedback_value"];
	        this.feedback_note = source["feedback_note"];
	    }
	}
	export class FeedbackResponse {
	    id: number;
	    target_type: string;
	    target_id: number;
	    target_index?: number;
	    feedback_value: string;
	    feedback_note: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new FeedbackResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.target_type = source["target_type"];
	        this.target_id = source["target_id"];
	        this.target_index = source["target_index"];
	        this.feedback_value = source["feedback_value"];
	        this.feedback_note = source["feedback_note"];
	        this.created_at = source["created_at"];
	    }
	}
	export class FinishTaskRequest {
	    finish_note: string;
	    finish_description: string;
	
	    static createFrom(source: any = {}) {
	        return new FinishTaskRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.finish_note = source["finish_note"];
	        this.finish_description = source["finish_description"];
	    }
	}
	export class GenerateSummaryResult {
	    summary_id: number;
	    content: string;
	    action_items?: any;
	
	    static createFrom(source: any = {}) {
	        return new GenerateSummaryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.summary_id = source["summary_id"];
	        this.content = source["content"];
	        this.action_items = source["action_items"];
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
	export class PlanRiskResponse {
	    date: string;
	    planned_total_minutes: number;
	    recent_avg_actual_minutes: number;
	    recent_active_days: number;
	    plan_ratio: number;
	    risk_level: string;
	    reason: string;
	    suggestions: string[];
	
	    static createFrom(source: any = {}) {
	        return new PlanRiskResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.planned_total_minutes = source["planned_total_minutes"];
	        this.recent_avg_actual_minutes = source["recent_avg_actual_minutes"];
	        this.recent_active_days = source["recent_active_days"];
	        this.plan_ratio = source["plan_ratio"];
	        this.risk_level = source["risk_level"];
	        this.reason = source["reason"];
	        this.suggestions = source["suggestions"];
	    }
	}
	export class Project {
	    id: number;
	    name: string;
	    description: string;
	    is_fixed: boolean;
	    category: string;
	    include_in_summary: boolean;
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
	        this.category = source["category"];
	        this.include_in_summary = source["include_in_summary"];
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
	    category: string;
	    include_in_summary: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProjectInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.is_fixed = source["is_fixed"];
	        this.category = source["category"];
	        this.include_in_summary = source["include_in_summary"];
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
	    action_items?: any;
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
	        this.action_items = source["action_items"];
	        this.created_at = source["created_at"];
	    }
	}
	export class UpdateCompletedTaskRequest {
	    finish_note: string;
	    finish_description: string;
	    actual_minutes_override?: number;
	    clear_actual_minutes_override?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpdateCompletedTaskRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.finish_note = source["finish_note"];
	        this.finish_description = source["finish_description"];
	        this.actual_minutes_override = source["actual_minutes_override"];
	        this.clear_actual_minutes_override = source["clear_actual_minutes_override"];
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

