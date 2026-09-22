export namespace common {
	
	export class GameVersion {
	
	
	    static createFrom(source: any = {}) {
	        return new GameVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class Language {
	    code: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Language(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	    }
	}

}

export namespace dto {
	
	export class TextRow {
	    index: number;
	    name?: string;
	    hash?: Record<string, string>;
	    text: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new TextRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.name = source["name"];
	        this.hash = source["hash"];
	        this.text = source["text"];
	    }
	}
	export class Metadata {
	    key: string;
	    row_count?: number;
	    id?: string;
	    is_dir?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Metadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.row_count = source["row_count"];
	        this.id = source["id"];
	        this.is_dir = source["is_dir"];
	    }
	}
	export class FileEntry {
	    metadata: Metadata;
	    rows: TextRow[];
	
	    static createFrom(source: any = {}) {
	        return new FileEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.metadata = this.convertValues(source["metadata"], Metadata);
	        this.rows = this.convertValues(source["rows"], TextRow);
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

export namespace services {
	
	export class EntrySummary {
	    id: string;
	    key: string;
	
	    static createFrom(source: any = {}) {
	        return new EntrySummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.key = source["key"];
	    }
	}

}

