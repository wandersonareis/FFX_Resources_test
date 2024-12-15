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
	export class ImportLocation {
	    TargetDirectory: string;
	    TargetDirectoryName: string;
	    IsExist: boolean;
	    TargetFile: string;
	    TargetPath: string;
	    TargetFileName: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportLocation(source);
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
	
	export class GameDataInfo {
	    file_path: string;
	    extract_location: locations.ExtractLocation;
	    translate_location: locations.TranslateLocation;
	    import_location: locations.ImportLocation;
	
	    static createFrom(source: any = {}) {
	        return new GameDataInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file_path = source["file_path"];
	        this.extract_location = this.convertValues(source["extract_location"], locations.ExtractLocation);
	        this.translate_location = this.convertValues(source["translate_location"], locations.TranslateLocation);
	        this.import_location = this.convertValues(source["import_location"], locations.ImportLocation);
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
	    data: GameDataInfo;
	    icon: string;
	    children: TreeNode[];
	
	    static createFrom(source: any = {}) {
	        return new TreeNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	        this.data = this.convertValues(source["data"], GameDataInfo);
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

