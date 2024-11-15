export namespace core {
	
	export class GameFiles {
	    name: string;
	    name_prefix: string;
	    size: number;
	    type: number;
	    extension: string;
	    parent: string;
	    is_dir: boolean;
	    full_path: string;
	    relative_path: string;
	    cloned_items: string[];
	
	    static createFrom(source: any = {}) {
	        return new GameFiles(source);
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
	        this.full_path = source["full_path"];
	        this.relative_path = source["relative_path"];
	        this.cloned_items = source["cloned_items"];
	    }
	}

}

export namespace interactions {
	
	export class ExtractLocation {
	    IsExist: boolean;
	    TargetFile: string;
	    TargetPath: string;
	    TargetFileName: string;
	    TargetDirectory: string;
	    TargetDirectoryName: string;
	
	    static createFrom(source: any = {}) {
	        return new ExtractLocation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.IsExist = source["IsExist"];
	        this.TargetFile = source["TargetFile"];
	        this.TargetPath = source["TargetPath"];
	        this.TargetFileName = source["TargetFileName"];
	        this.TargetDirectory = source["TargetDirectory"];
	        this.TargetDirectoryName = source["TargetDirectoryName"];
	    }
	}
	export class ImportLocation {
	    IsExist: boolean;
	    TargetFile: string;
	    TargetPath: string;
	    TargetFileName: string;
	    TargetDirectory: string;
	    TargetDirectoryName: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportLocation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.IsExist = source["IsExist"];
	        this.TargetFile = source["TargetFile"];
	        this.TargetPath = source["TargetPath"];
	        this.TargetFileName = source["TargetFileName"];
	        this.TargetDirectory = source["TargetDirectory"];
	        this.TargetDirectoryName = source["TargetDirectoryName"];
	    }
	}
	export class TranslateLocation {
	    IsExist: boolean;
	    TargetFile: string;
	    TargetPath: string;
	    TargetFileName: string;
	    TargetDirectory: string;
	    TargetDirectoryName: string;
	
	    static createFrom(source: any = {}) {
	        return new TranslateLocation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.IsExist = source["IsExist"];
	        this.TargetFile = source["TargetFile"];
	        this.TargetPath = source["TargetPath"];
	        this.TargetFileName = source["TargetFileName"];
	        this.TargetDirectory = source["TargetDirectory"];
	        this.TargetDirectoryName = source["TargetDirectoryName"];
	    }
	}
	export class GameDataInfo {
	    game_data: core.GameFiles;
	    extract_location: ExtractLocation;
	    translate_location: TranslateLocation;
	    import_location: ImportLocation;
	
	    static createFrom(source: any = {}) {
	        return new GameDataInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.game_data = this.convertValues(source["game_data"], core.GameFiles);
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

