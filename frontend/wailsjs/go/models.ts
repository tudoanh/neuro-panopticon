export namespace agent {
	
	export class ChatResponse {
	    content: string;
	    tool_calls?: models.ToolCall[];
	
	    static createFrom(source: any = {}) {
	        return new ChatResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.content = source["content"];
	        this.tool_calls = this.convertValues(source["tool_calls"], models.ToolCall);
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

export namespace config {
	
	export class SecurityConfig {
	    allow_remediation: boolean;
	    allowed_remediations: string[];
	    scan_interval_seconds: number;
	
	    static createFrom(source: any = {}) {
	        return new SecurityConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.allow_remediation = source["allow_remediation"];
	        this.allowed_remediations = source["allowed_remediations"];
	        this.scan_interval_seconds = source["scan_interval_seconds"];
	    }
	}
	export class LLMConfig {
	    backend: string;
	    ollama_url: string;
	    ollama_model: string;
	    cloud_api_key?: string;
	    cloud_url?: string;
	    cloud_model?: string;
	    max_tokens: number;
	    temperature: number;
	
	    static createFrom(source: any = {}) {
	        return new LLMConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.backend = source["backend"];
	        this.ollama_url = source["ollama_url"];
	        this.ollama_model = source["ollama_model"];
	        this.cloud_api_key = source["cloud_api_key"];
	        this.cloud_url = source["cloud_url"];
	        this.cloud_model = source["cloud_model"];
	        this.max_tokens = source["max_tokens"];
	        this.temperature = source["temperature"];
	    }
	}
	export class AppConfig {
	    llm: LLMConfig;
	    security: SecurityConfig;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.llm = this.convertValues(source["llm"], LLMConfig);
	        this.security = this.convertValues(source["security"], SecurityConfig);
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

export namespace models {
	
	export class Finding {
	    id: string;
	    title: string;
	    description: string;
	    severity: string;
	    category: string;
	    source: string;
	    remediation?: string;
	    // Go type: time
	    timestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new Finding(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.severity = source["severity"];
	        this.category = source["category"];
	        this.source = source["source"];
	        this.remediation = source["remediation"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
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
	export class SecurityScore {
	    score: number;
	    max_score: number;
	    findings: number;
	    critical: number;
	    high: number;
	    medium: number;
	    low: number;
	    // Go type: time
	    last_updated: any;
	
	    static createFrom(source: any = {}) {
	        return new SecurityScore(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.score = source["score"];
	        this.max_score = source["max_score"];
	        this.findings = source["findings"];
	        this.critical = source["critical"];
	        this.high = source["high"];
	        this.medium = source["medium"];
	        this.low = source["low"];
	        this.last_updated = this.convertValues(source["last_updated"], null);
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
	export class SystemStatus {
	    cpu_percent: number;
	    memory_used_gb: number;
	    memory_total_gb: number;
	    network_in_bytes: number;
	    network_out_bytes: number;
	    security_score: SecurityScore;
	    agent_ready: boolean;
	    llm_backend: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpu_percent = source["cpu_percent"];
	        this.memory_used_gb = source["memory_used_gb"];
	        this.memory_total_gb = source["memory_total_gb"];
	        this.network_in_bytes = source["network_in_bytes"];
	        this.network_out_bytes = source["network_out_bytes"];
	        this.security_score = this.convertValues(source["security_score"], SecurityScore);
	        this.agent_ready = source["agent_ready"];
	        this.llm_backend = source["llm_backend"];
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
	export class ToolCall {
	    id: string;
	    name: string;
	    args: string;
	    result?: string;
	    status: string;
	    duration_ms?: number;
	
	    static createFrom(source: any = {}) {
	        return new ToolCall(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.args = source["args"];
	        this.result = source["result"];
	        this.status = source["status"];
	        this.duration_ms = source["duration_ms"];
	    }
	}

}

