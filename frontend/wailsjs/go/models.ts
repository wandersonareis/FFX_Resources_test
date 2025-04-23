export namespace core {
	
	export class SpiraFileInfo {
	    name: string;
	    name_prefix: string;
	    type: number;
	    size: number;
	    extension: string;
	    entry_path: string;
	    parent: string;
	    is_dir: boolean;
	    cloned_items: string[];
	    path: string;
	    relative_path: string;
	
	    static createFrom(source: any = {}) {
	        return new SpiraFileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.name_prefix = source["name_prefix"];
	        this.type = source["type"];
	        this.size = source["size"];
	        this.extension = source["extension"];
	        this.entry_path = source["entry_path"];
	        this.parent = source["parent"];
	        this.is_dir = source["is_dir"];
	        this.cloned_items = source["cloned_items"];
	        this.path = source["path"];
	        this.relative_path = source["relative_path"];
	    }
	}

}

export namespace fileFormats {
	
	export class TreeNodeData {
	    source?: core.SpiraFileInfo;
	    extract_location?: locations.ExtractLocation;
	    translate_location?: locations.TranslateLocation;
	
	    static createFrom(source: any = {}) {
	        return new TreeNodeData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = this.convertValues(source["source"], core.SpiraFileInfo);
	        this.extract_location = this.convertValues(source["extract_location"], locations.ExtractLocation);
	        this.translate_location = this.convertValues(source["translate_location"], locations.TranslateLocation);
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

export namespace locations {
	
	export class ExtractLocation {
	    TargetDirectory: string;
	    TargetDirectoryName: string;
	    IsExist: boolean;
	    TargetFile: string;
	    TargetPath: string;
	    TargetFileName: string;
	
	    static createFrom(source: any = {}) {
	        return new ExtractLocation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TargetDirectory = source["TargetDirectory"];
	        this.TargetDirectoryName = source["TargetDirectoryName"];
	        this.IsExist = source["IsExist"];
	        this.TargetFile = source["TargetFile"];
	        this.TargetPath = source["TargetPath"];
	        this.TargetFileName = source["TargetFileName"];
	    }
	}
	export class TranslateLocation {
	    TargetDirectory: string;
	    TargetDirectoryName: string;
	    IsExist: boolean;
	    TargetFile: string;
	    TargetPath: string;
	    TargetFileName: string;
	
	    static createFrom(source: any = {}) {
	        return new TranslateLocation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TargetDirectory = source["TargetDirectory"];
	        this.TargetDirectoryName = source["TargetDirectoryName"];
	        this.IsExist = source["IsExist"];
	        this.TargetFile = source["TargetFile"];
	        this.TargetPath = source["TargetPath"];
	        this.TargetFileName = source["TargetFileName"];
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

