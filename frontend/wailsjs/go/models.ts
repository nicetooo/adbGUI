export namespace main {
	
	export class AppPackage {
	    name: string;
	    label: string;
	    icon: string;
	    type: string;
	    state: string;
	    versionName: string;
	    versionCode: string;
	    minSdkVersion: string;
	    targetSdkVersion: string;
	    permissions: string[];
	    activities: string[];
	    launchableActivities: string[];
	
	    static createFrom(source: any = {}) {
	        return new AppPackage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.label = source["label"];
	        this.icon = source["icon"];
	        this.type = source["type"];
	        this.state = source["state"];
	        this.versionName = source["versionName"];
	        this.versionCode = source["versionCode"];
	        this.minSdkVersion = source["minSdkVersion"];
	        this.targetSdkVersion = source["targetSdkVersion"];
	        this.permissions = source["permissions"];
	        this.activities = source["activities"];
	        this.launchableActivities = source["launchableActivities"];
	    }
	}
	export class Device {
	    id: string;
	    state: string;
	    model: string;
	    brand: string;
	
	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.state = source["state"];
	        this.model = source["model"];
	        this.brand = source["brand"];
	    }
	}
	export class DeviceInfo {
	    model: string;
	    brand: string;
	    manufacturer: string;
	    androidVer: string;
	    sdk: string;
	    abi: string;
	    serial: string;
	    resolution: string;
	    density: string;
	    cpu: string;
	    memory: string;
	    props: {[key: string]: string};
	
	    static createFrom(source: any = {}) {
	        return new DeviceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.brand = source["brand"];
	        this.manufacturer = source["manufacturer"];
	        this.androidVer = source["androidVer"];
	        this.sdk = source["sdk"];
	        this.abi = source["abi"];
	        this.serial = source["serial"];
	        this.resolution = source["resolution"];
	        this.density = source["density"];
	        this.cpu = source["cpu"];
	        this.memory = source["memory"];
	        this.props = source["props"];
	    }
	}
	export class FileInfo {
	    name: string;
	    size: number;
	    mode: string;
	    modTime: string;
	    isDir: boolean;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.size = source["size"];
	        this.mode = source["mode"];
	        this.modTime = source["modTime"];
	        this.isDir = source["isDir"];
	        this.path = source["path"];
	    }
	}
	export class ScrcpyConfig {
	    maxSize: number;
	    bitRate: number;
	    maxFps: number;
	    stayAwake: boolean;
	    turnScreenOff: boolean;
	    noAudio: boolean;
	    alwaysOnTop: boolean;
	    showTouches: boolean;
	    fullscreen: boolean;
	    readOnly: boolean;
	    powerOffOnClose: boolean;
	    windowBorderless: boolean;
	    videoCodec: string;
	    audioCodec: string;
	    recordPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ScrcpyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maxSize = source["maxSize"];
	        this.bitRate = source["bitRate"];
	        this.maxFps = source["maxFps"];
	        this.stayAwake = source["stayAwake"];
	        this.turnScreenOff = source["turnScreenOff"];
	        this.noAudio = source["noAudio"];
	        this.alwaysOnTop = source["alwaysOnTop"];
	        this.showTouches = source["showTouches"];
	        this.fullscreen = source["fullscreen"];
	        this.readOnly = source["readOnly"];
	        this.powerOffOnClose = source["powerOffOnClose"];
	        this.windowBorderless = source["windowBorderless"];
	        this.videoCodec = source["videoCodec"];
	        this.audioCodec = source["audioCodec"];
	        this.recordPath = source["recordPath"];
	    }
	}

}

