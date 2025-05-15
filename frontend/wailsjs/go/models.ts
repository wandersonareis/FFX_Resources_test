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
	    version: number;
	
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
	        this.version = source["version"];
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

