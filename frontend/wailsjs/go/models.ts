export namespace lib {
	
	export class ExtractLocation {
	    TargetDirectoryName: string;
	    TargetDirectory: string;
	    TargetFile: string;
	    TargetPath: string;
	    IsExist: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ExtractLocation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TargetDirectoryName = source["TargetDirectoryName"];
	        this.TargetDirectory = source["TargetDirectory"];
	        this.TargetFile = source["TargetFile"];
	        this.TargetPath = source["TargetPath"];
	        this.IsExist = source["IsExist"];
	    }
	}
	export class ImportLocation {
	    TargetDirectoryName: string;
	    TargetDirectory: string;
	    TargetFile: string;
	    TargetPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportLocation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TargetDirectoryName = source["TargetDirectoryName"];
	        this.TargetDirectory = source["TargetDirectory"];
	        this.TargetFile = source["TargetFile"];
	        this.TargetPath = source["TargetPath"];
	    }
	}
	export class TranslateLocation {
	    TargetDirectoryName: string;
	    TargetDirectory: string;
	    TargetFile: string;
	    TargetPath: string;
	    IsExist: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TranslateLocation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TargetDirectoryName = source["TargetDirectoryName"];
	        this.TargetDirectory = source["TargetDirectory"];
	        this.TargetFile = source["TargetFile"];
	        this.TargetPath = source["TargetPath"];
	        this.IsExist = source["IsExist"];
	    }
	}
	export class FileInfo {
	    name: string;
	    name_prefix: string;
	    size: number;
	    type: number;
	    extension: string;
	    parent: string;
	    is_dir: boolean;
	    absolute_path: string;
	    relative_path: string;
	    extract_location: ExtractLocation;
	    translate_location: TranslateLocation;
	    import_location: ImportLocation;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.name_prefix = source["name_prefix"];
	        this.size = source["size"];
	        this.type = source["type"];
	        this.extension = source["extension"];
	        this.parent = source["parent"];
	        this.is_dir = source["is_dir"];
	        this.absolute_path = source["absolute_path"];
	        this.relative_path = source["relative_path"];
	        this.extract_location = this.convertValues(source["extract_location"], ExtractLocation);
	        this.translate_location = this.convertValues(source["translate_location"], TranslateLocation);
	        this.import_location = this.convertValues(source["import_location"], ImportLocation);
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
	
	
	export class TreeNode {
	    key: string;
	    label: string;
	    data: FileInfo;
	    icon: string;
	    children: TreeNode[];
	
	    static createFrom(source: any = {}) {
	        return new TreeNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	        this.data = this.convertValues(source["data"], FileInfo);
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

export namespace main {
	
	export class AppConfig {
	    OriginalDirectory: string;
	    ExtractDirectory: string;
	    TranslateDirectory: string;
	    // Go type: lib
	    GameLocation: any;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.OriginalDirectory = source["OriginalDirectory"];
	        this.ExtractDirectory = source["ExtractDirectory"];
	        this.TranslateDirectory = source["TranslateDirectory"];
	        this.GameLocation = this.convertValues(source["GameLocation"], null);
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

