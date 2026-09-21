export namespace common {
	
	export class GameVersion {
	
	
	    static createFrom(source: any = {}) {
	        return new GameVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
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

export namespace fileFormats {
	
	export class TreeNodeData {
	    source: models.SpiraFileInfo;
	    extract_location: any;
	    translate_location: any;
	
	    static createFrom(source: any = {}) {
	        return new TreeNodeData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = this.convertValues(source["source"], models.SpiraFileInfo);
	        this.extract_location = source["extract_location"];
	        this.translate_location = source["translate_location"];
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
	
	export class SpiraFileInfo {
	    name: string;
	    name_prefix: string;
	    extension: string;
	    is_dir: boolean;
	    cloned_items: string[];
	    path: string;
	    parent: string;
	    relative_path: string;
	    size: number;
	    type: number;
	    version: common.GameVersion;
	
	    static createFrom(source: any = {}) {
	        return new SpiraFileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.name_prefix = source["name_prefix"];
	        this.extension = source["extension"];
	        this.is_dir = source["is_dir"];
	        this.cloned_items = source["cloned_items"];
	        this.path = source["path"];
	        this.parent = source["parent"];
	        this.relative_path = source["relative_path"];
	        this.size = source["size"];
	        this.type = source["type"];
	        this.version = this.convertValues(source["version"], common.GameVersion);
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

export namespace spira {
	
	export class TreeNode {
	    key: string;
	    label: string;
	    data?: fileFormats.TreeNodeData;
	    icon: string;
	    children: TreeNode[];
	
	    static createFrom(source: any = {}) {
	        return new TreeNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	        this.data = this.convertValues(source["data"], fileFormats.TreeNodeData);
	        this.icon = source["icon"];
	        this.children = this.convertValues(source["children"], TreeNode);
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

