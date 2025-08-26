export namespace auth {
	
	export class LoginRequest {
	    username: string;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new LoginRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.password = source["password"];
	    }
	}
	export class User {
	    id: string;
	    username: string;
	    email: string;
	    password_hash: string;
	    salt: string;
	    created_at: string;
	    last_login: string;
	    is_active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new User(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.email = source["email"];
	        this.password_hash = source["password_hash"];
	        this.salt = source["salt"];
	        this.created_at = source["created_at"];
	        this.last_login = source["last_login"];
	        this.is_active = source["is_active"];
	    }
	}
	export class LoginResponse {
	    success: boolean;
	    message: string;
	    session_id?: string;
	    user?: User;
	
	    static createFrom(source: any = {}) {
	        return new LoginResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.session_id = source["session_id"];
	        this.user = this.convertValues(source["user"], User);
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
	export class RegisterRequest {
	    username: string;
	    email: string;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new RegisterRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.email = source["email"];
	        this.password = source["password"];
	    }
	}
	export class Session {
	    id: string;
	    user_id: string;
	    username: string;
	    created_at: string;
	    expires_at: string;
	    is_active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Session(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.user_id = source["user_id"];
	        this.username = source["username"];
	        this.created_at = source["created_at"];
	        this.expires_at = source["expires_at"];
	        this.is_active = source["is_active"];
	    }
	}

}

export namespace engine {
	
	export class BackupInfo {
	    name: string;
	    database: string;
	    // Go type: time
	    timestamp: any;
	    size: number;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new BackupInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.database = source["database"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.size = source["size"];
	        this.path = source["path"];
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
	export class CollectionStats {
	    name: string;
	    document_count: number;
	    index_count: number;
	    avg_doc_size: number;
	    field_types: Record<string, string>;
	    index_efficiency: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new CollectionStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.document_count = source["document_count"];
	        this.index_count = source["index_count"];
	        this.avg_doc_size = source["avg_doc_size"];
	        this.field_types = source["field_types"];
	        this.index_efficiency = source["index_efficiency"];
	    }
	}
	export class DatabaseStats {
	    name: string;
	    collections: number;
	    total_documents: number;
	    total_indexes: number;
	    size_on_disk: number;
	    collection_stats: Record<string, CollectionStats>;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.collections = source["collections"];
	        this.total_documents = source["total_documents"];
	        this.total_indexes = source["total_indexes"];
	        this.size_on_disk = source["size_on_disk"];
	        this.collection_stats = this.convertValues(source["collection_stats"], CollectionStats, true);
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
	export class ImportResult {
	    imported: number;
	    skipped: number;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new ImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imported = source["imported"];
	        this.skipped = source["skipped"];
	        this.errors = source["errors"];
	    }
	}

}

export namespace service {
	
	export class SortOption {
	    field: string;
	    ascending: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SortOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.field = source["field"];
	        this.ascending = source["ascending"];
	    }
	}
	export class QueryFilter {
	    field: string;
	    operator: string;
	    value: any;
	
	    static createFrom(source: any = {}) {
	        return new QueryFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.field = source["field"];
	        this.operator = source["operator"];
	        this.value = source["value"];
	    }
	}
	export class AdvancedQueryRequest {
	    database: string;
	    collection: string;
	    filters: QueryFilter[];
	    sort?: SortOption;
	    limit: number;
	    skip: number;
	
	    static createFrom(source: any = {}) {
	        return new AdvancedQueryRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.database = source["database"];
	        this.collection = source["collection"];
	        this.filters = this.convertValues(source["filters"], QueryFilter);
	        this.sort = this.convertValues(source["sort"], SortOption);
	        this.limit = source["limit"];
	        this.skip = source["skip"];
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
	export class BackupRequest {
	    database: string;
	    backup_name: string;
	
	    static createFrom(source: any = {}) {
	        return new BackupRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.database = source["database"];
	        this.backup_name = source["backup_name"];
	    }
	}
	export class CollectionInfo {
	    name: string;
	    document_count: number;
	    indexes: string[];
	
	    static createFrom(source: any = {}) {
	        return new CollectionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.document_count = source["document_count"];
	        this.indexes = source["indexes"];
	    }
	}
	export class DatabaseInfo {
	    name: string;
	    collections: string[];
	
	    static createFrom(source: any = {}) {
	        return new DatabaseInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.collections = source["collections"];
	    }
	}
	export class DeleteRequest {
	    database: string;
	    collection: string;
	    id: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.database = source["database"];
	        this.collection = source["collection"];
	        this.id = source["id"];
	    }
	}
	export class DocumentResponse {
	    id: string;
	    data: Record<string, any>;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new DocumentResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.data = source["data"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class ExportRequest {
	    database: string;
	    collection: string;
	    format: string;
	    query?: AdvancedQueryRequest;
	    file_path: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.database = source["database"];
	        this.collection = source["collection"];
	        this.format = source["format"];
	        this.query = this.convertValues(source["query"], AdvancedQueryRequest);
	        this.file_path = source["file_path"];
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
	export class ImportRequest {
	    database: string;
	    collection: string;
	    format: string;
	    file_path: string;
	    create_collection: boolean;
	    overwrite_data: boolean;
	    id_field: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.database = source["database"];
	        this.collection = source["collection"];
	        this.format = source["format"];
	        this.file_path = source["file_path"];
	        this.create_collection = source["create_collection"];
	        this.overwrite_data = source["overwrite_data"];
	        this.id_field = source["id_field"];
	    }
	}
	export class InsertRequest {
	    database: string;
	    collection: string;
	    id: string;
	    data: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new InsertRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.database = source["database"];
	        this.collection = source["collection"];
	        this.id = source["id"];
	        this.data = source["data"];
	    }
	}
	
	export class QueryRequest {
	    database: string;
	    collection: string;
	    field: string;
	    value: any;
	
	    static createFrom(source: any = {}) {
	        return new QueryRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.database = source["database"];
	        this.collection = source["collection"];
	        this.field = source["field"];
	        this.value = source["value"];
	    }
	}
	export class RestoreRequest {
	    backup_path: string;
	    new_db_name: string;
	
	    static createFrom(source: any = {}) {
	        return new RestoreRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.backup_path = source["backup_path"];
	        this.new_db_name = source["new_db_name"];
	    }
	}
	
	export class UpdateRequest {
	    database: string;
	    collection: string;
	    id: string;
	    data: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new UpdateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.database = source["database"];
	        this.collection = source["collection"];
	        this.id = source["id"];
	        this.data = source["data"];
	    }
	}

}

