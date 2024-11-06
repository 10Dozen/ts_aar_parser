var eStyle = {
	"headerStatus": {
		"default": {
			"text": "Open RPT file to convert...",
			"bgColor": "#5a5a5a"			
		},
		"suggest": {
			"text": "Select AAR to parse",
			"bgColor": "rgb(155, 195, 78)"
		},
		"success": {
			"text": "Choose report from the list to convert!",
			"bgColor": "rgb(155, 195, 78)"
		},
		"failedEmpty": {
			"text": "File is empty/Failed to read!",
			"bgColor": "#AF4E4E"
		},
		"failedWrong": {
			"text": "File does not contain AAR data!",
			"bgColor": "#AF4E4E"
		}
	}
};

var statusElement = {
	"text": "#header-status-text",
	"bar": "#header-status"
};

var enableInterpolation = false;
var zipAARFile = true;
var rptData = [];
var reportGuid = "";
var allParsedMeta = [];
var aarData;

const windowsFilenameEscapeSymbols = [
	/:/g,
	/\*/g,
	/\?/g,
	/</g,
	/>/g,
	/\|/g,
	/"/g,
	/\\/g,
	/\//g,
	/ /g
];

function normalize(s) {
	return s.replace(/""/g, '"')
};

function normalizeFilename(s) {
	windowsFilenameEscapeSymbols.forEach(e => s = s.replace(e, "_"))
	return s
}

function updateHeaderStatus(mode) {
	$( statusElement.text ).html( eStyle.headerStatus[mode].text );
	$( statusElement.bar ).css( "background-color",  eStyle.headerStatus[mode].bgColor )
}

function resetForm(wipe = false) {
	$( "#header-back-btn").css("display", "none");
	$( "#result-form" ).css( "top", "-1000px" );
	$( "#mission-desc" ).val("");
	$( "#mission-date" ).val("");
	$( "#output" ).val("");
	
	$( "#report-selector" ).css( "top", "-1000px" );
	$( "#report-selector > ul" ).html( "" );
	
	if (wipe) {
		updateHeaderStatus( "default" );

		rptData = [];
		allParsedMeta = [];
	}
	reportGuid = "";
	aarData = [];
}

// Open file
var openFile = function(event) { 
	resetForm();
	toggleProgressView(true);
	updateProgressView("Opening...", "Please, wait until file read.");
	
	setTimeout( readFile(event), 200 );
}
var openConfigFile = function(event) {
	console.log('Config File Selected');
	var input = event.target;
    var configReader = new FileReader();

	configFilename = configUploader.files[0].name;
	configReader.onload = function(){
		console.log("READING");
		var configData = configReader.result;
		if (configData.length > 0) {
			console.log( "Read!");
			var newConfig = ( AARFileDetails.configLine.replace(/\n/g, '') ).slice(0,-1);

			eval(configData);	// this creates aarConfig variable with configs
			aarConfig.unshift(JSON.parse(normalize(newConfig)));

			var stringified = JSON.stringify( aarConfig );
			stringified = stringified.replace(/,/g, ',\n').replace(/{/g, '{\n').replace(/}/g, '\n}');
			stringified = stringified.replace(/"date"/g, '	"date"').replace(/"terrain"/g, '	"terrain"').replace(/"link"/g, '	"link"').replace(/"title"/g, '	"title"');

			saveConfigFile("aarConfig = " + stringified + ";");
        } else {
        	console.log( "Not Read!");
        }
	}
	configReader.readAsText(input.files[0]);
};

function readFile(event) {
	let readoutMetadata = function (s) {
		let elements = s.replace("{","").replace("}","").split(/[:,]/).map((e) => e.trim().replaceAll('""',''));
		let metadataObj = {}
		for (let i = 0; i < elements.length; i += 2) {
			metadataObj[elements[i]] = elements[i+1];
		}

		return metadataObj
	}

	let input = event.target;
	let reader = new FileReader();

	let filename = uploader.files[0].name;
	let filenameElements = filename.split("_");
	let aarFileDate =  (filenameElements[1]).toLowerCase() == "x64" ? filenameElements[2] : filenameElements[1];


	reader.onload = function(){
		let lines = reader.result.split("\n");
		if (lines.length == 0) {
			console.log( "Empty!" );
			updateHeaderStatus( "failedEmpty" );
			return
		};

		let metaTestPattern = /<meta><core>/i
		let metaMatchPattern = /(.*)<AAR-.*><meta><core>(.*)<\/core><\/meta><\/AAR-.*>/i
		let aarTestPattern = /<AAR-.*>/i

		// Read all AAR related RPT data
		rptData = lines.filter((l) => aarTestPattern.test(l));

		// Read metadata lines
		let metadataLines = rptData.filter((l) => metaTestPattern.test(l))
		if (metadataLines.length == 0) {
			console.log( "No AAR found!" );
			updateHeaderStatus( "failedWrong" );
			return
		}

		console.log( "Read!");
		allParsedMeta = [];
		metadataLines.forEach((l) => {
			let matchGroups = l.match(metaMatchPattern)
			// Check for
			let metadata = readoutMetadata(matchGroups[2])
			metadata.logTime = matchGroups[1].slice(0,-2);
			metadata.logDate = aarFileDate;
			console.log(metadata)

			allParsedMeta.push(metadata);
		})

		setTimeout( showAARParseSuggestion(), 500 );

		updateHeaderStatus( "success" );
		$( "#report-selector" ).css( "top", "75px" );
	};
	reader.readAsText(input.files[0]);
};

function showAARParseSuggestion () {
	resetForm(false);
	$( "#report-selector" ).css( "top", "75px" );
	updateHeaderStatus( "suggest" );
	toggleProgressView(false)

	let addButton = function (meta) {
		$( "#report-selector > ul" ).append(
			`<li onClick='chooseReportToConvert(\"${meta.guid}\");'>${meta.island} ▸ ${meta.logTime} ▸ ${meta.name}</li>`
		);
	}
	allParsedMeta.forEach(e => addButton(e));
}

function chooseReportToConvert(guid) {	
	reportGuid = guid;
	
	$( "#report-selector > ul" ).html("");
	$( "#report-selector" ).css("top", "-1000px");
	convertInit();
};

function convertInit() {
	toggleProgressView(true);
	updateProgressView( "Converting...","Please, wait until conversion is over." );
	setTimeout( convertToAAR, 500 );
}

function toggleProgressView(on) {
	$( "#progress-header" ).html( "" );
	$( "#progress-status" ).html( "" );
	if (on) {
		$( "#progress-viewer" ).css("top", "150px");
	} else {
		$( "#progress-viewer" ).css("top", "-1000px");
	}
}

function updateProgressView(header, text) {
	$( "#progress-header" ).html( header );
	$( "#progress-status" ).append( text + "<br />" );	
}

function convertToAAR() {
	console.log(reportGuid);
	
	aarData = {
		"metadata": {
			"island": "",
			"name": "",
			"time": 0,
			"date": "",
			"desc": "",
			"players": [],
			"objects": {
				"units": [],
				"vehs": []					
			}
		},
		"timeline": []
	};
	
	let consoleMsgEnabled = true;
	let consoleDebugEnabled = false;
	function logMsg(t) { if (consoleMsgEnabled) { console.log( t ) }};
	function logDebug(t) { if (consoleDebugEnabled) { console.log( t ) }};
	
	let metadataCore = allParsedMeta.filter((e)=>e.guid == reportGuid)[0];
	
	logMsg( "Metadata: Core [ Processing ]" );
	aarData.metadata.island = metadataCore.island;
	aarData.metadata.name = metadataCore.name;
	aarData.metadata.desc = metadataCore.summary;
	aarData.metadata.date = metadataCore.logDate;
	logMsg( "Metadata: Core [ OK ]" );

	let re = new RegExp( `<AAR-${reportGuid}>`, "i" )
	let aarLines = rptData.filter((e)=>re.test(e))

	let actorMetaTestPattern = /<meta><(veh|unit)>/i
	let actorMetaPattern = /<meta><(unit|veh)>(.*)<\/(unit|veh)><\/meta>/i
	let eventTestPattern = /<(\d+)><(unit|veh|av)>/i
	let eventPattern = /<(\d+)><(unit|veh|av)>(.*)<\/(unit|veh|av)>/i

	logMsg( "Objects [ Processing ]" );	
	for (let i = 0; i < aarLines.length; i++) {
		let line = aarLines[i];

		// Check for unit's metadata
		if (actorMetaTestPattern.test(line)) {
			let matchGroups = line.match(actorMetaPattern)
			if (matchGroups == null) continue;
			[ _, actorType, actorMetadata ] = matchGroups
			actorMetadata = JSON.parse(normalize(actorMetadata))

			if (actorType == "veh") {
				aarData.metadata.objects.vehs.push(actorMetadata.vehMeta)
			} else {
				aarData.metadata.objects.units.push(actorMetadata.unitMeta)
				if (actorMetadata.unitMeta[3] > 0) {
					let playerData = [
						actorMetadata.unitMeta[1],
						actorMetadata.unitMeta[2]
					]
					let alreadyReported = aarData.metadata.players.some((e)=> {
						return e[0] == playerData[0] && e[1] == playerData[1]
					})
					if (alreadyReported) continue;
					(aarData.metadata.players).push(playerData);
				}
			}
			continue;
		}

		// Check for event data
		if (eventTestPattern.test(line)) {
			let matchGroups = line.match(eventPattern);
			if (matchGroups == null) continue;
			[ _, timelabel, eventType, eventData ] = matchGroups;

			let timelineEvents = aarData.timeline[timelabel]
			if (timelineEvents == null) {
				timelineEvents = [ [], [], [] ];
				aarData.timeline[timelabel] = timelineEvents;
			}

			eventData = JSON.parse(normalize(eventData));
			eventTypeId = ["unit","veh","av"].indexOf(eventType)

			timelineEvents[eventTypeId].push(eventData);
			continue;
		}
	};
	
	logMsg( "Objects [ OK ]" );
	
	logMsg( "Check timeline [ Processing ]" );
	for (var i=0; i < aarData.timeline.length; i++) {
		try { 
			aarData.timeline[i].length > 0
		} catch(e) { 
			aarData.timeline[i] = [[],[],[]];
		};
	};
	logMsg( "Check timeline [ OK ]" );

	if (enableInterpolation) {
	logMsg( "Timeline: Interpolating Transitions of Units [ Processing ]" );
	/*
		For each UNIT check all timelines.
			If there are no data for timeline 
			- select last time when data exists then find time when data exists next (or end of time)
			- then for each time between last known and latest time - calculate/interpolate values of position
			- add this data to timeline
			- check for another range	
	*/	
	for (var m = 0; m < 2; m++) {		
		var unitList, unitTypeId;

		if (m == 0) {
			unitList = aarData.metadata.objects.units;
			unitTypeId = 0;
		} else {			
			unitList = aarData.metadata.objects.vehs;
			unitTypeId = 1;
		}		
		
		for (var i = 0; i < unitList.length; i++ ) {
			var unitId = unitList[i][0];			
			logDebug("INTERPOLATION FOR UNIT " + unitId);

			var lastKnownTimestamp = 0;
			var lastKnown = [];
			var actualKnownTimestamp = 0;
			var actualKnown = [];
			var stepsToInterpolate = [];
			
			// **************
			// Looks like if we start to seek for timeframes not from 0 - it will easily avoid interpolation of afterspawned units
			// But it will cause few second of lag of bots, but who carse
			// ***********
			// For each Second
			for (var j = 1; j < aarData.timeline.length; j++) {
				// For each Unit per Timeline item
				for (var k = 0; k < aarData.timeline[j][unitTypeId].length; k++) {
					if (aarData.timeline[j][unitTypeId][k][0] == unitId || j == (aarData.timeline.length - 1)) {
						logDebug("Time " + j + " unit data is here! " + aarData.timeline[j][unitTypeId][k]);

						if (lastKnown.length == 0) {
							logDebug("Start of Interpol Range");
							lastKnown = aarData.timeline[j][unitTypeId][k];
							lastKnownTimestamp = j;
						} else {
							if (actualKnown.length == 0) {
								logDebug("End of Interpol Range");
								actualKnown = aarData.timeline[j][unitTypeId][k];
								actualKnownTimestamp = j;
							}
						}
					}
				}
	
				if (lastKnown.length > 0 && actualKnown.length == 0) {
					logDebug("Adding time to empty");
					stepsToInterpolate.push(j);
				} else {
					if (lastKnown.length > 0 && actualKnown.length > 0 && (  lastKnownTimestamp != actualKnownTimestamp ) ) {
						logDebug("Interpolation ( @" + lastKnown[1] + ", @" + actualKnown[1] + ", @" + stepsToInterpolate.length + ")");

						var posxSteps = interpolateValues( lastKnown[1], actualKnown[1], stepsToInterpolate.length );
						var posySteps = interpolateValues( lastKnown[2], actualKnown[2], stepsToInterpolate.length );
						var dirSteps = lastKnown[3];
						//var dirSteps = interpolateValues( lastKnown[3], actualKnown[3], stepsToInterpolate.length );

						logDebug("Interpolation... | X: " + posxSteps + " || Y: " + posySteps + " || DIR: " + dirSteps);								
						// Start with 1, because 0 is equal to lastKnown value and we may not update it
						for (var l = 1; l < stepsToInterpolate.length; l++ ) {
							logDebug( "Time (" + stepsToInterpolate[l] + ")" + [unitId, posxSteps[l], posySteps[l], dirSteps, lastKnown[4], lastKnown[5] ] );
							if (m == 0) {
								if (lastKnown[4] == 1) {
									( aarData.timeline[stepsToInterpolate[l]][unitTypeId] ).push( [
										unitId
										, posxSteps[l]
										, posySteps[l]
										, dirSteps
										, lastKnown[4]
										, lastKnown[5]
									]  );
								}
							} else {
								( aarData.timeline[stepsToInterpolate[l]][unitTypeId] ).push( [unitId, posxSteps[l], posySteps[l], dirSteps, lastKnown[4], lastKnown[5], lastKnown[6]]  );
							}
						}
						
						lastKnown = [];
						actualKnown = [];
						stepsToInterpolate = [];								
						// Step back to get new start interpolation range
						j = j - 1;
					}
				}
			}
		}
	}
	
	logMsg( "Timeline: Interpolating Transitions of Units [ OK ]" );
	};

	aarData.metadata.time = (aarData.timeline).length - 2;

	logMsg( "Creating form" );
	$( "#player-list" ).html( "" );
	$( "#mission-time" ).html( getTimeLabel(aarData.metadata.time) );
	
	for (var i = 0; i < aarData.metadata.players.length; i++) {
		var color;
		switch (aarData.metadata.players[i][1]) {
			case "blufor": color = "RGB(0,77,152)"; break;
			case "opfor": color = "RGB(127,0,0)"; break;
			case "indep": color = "RGB(0,127,0)"; break;
			case "civ": color = "RGB(102,0,127)"; break;
		};
		$( "#player-list" ).append( "<li class='player-side-icon' style='padding: 2px 4px; background-color: " + color + "'>" + aarData.metadata.players[i][0] + "</li>" );				
	}

	logMsg("Done!");
	toggleProgressView(false);
	$( "#result-form" ).css( "top", "75px" );
	$( "#header-status-text" ).html( "Converted!" );
	$( ".dl-2 > input, textarea" ).removeAttr( "disabled" );

	AARFileDetails = new AARFileDetailsBase();
};

// Interpolate values
function interpolateValues(min,max,steps) {
	var output = [];
	var delta = (max - min)/steps;
	for (var i = 0; i < steps; i++) { output.push( Math.round( min + i*delta ) ); }

	return output			
}
			
// Time label
function getTimeLabel(t) {
	var time = t;
	var timeHours = time / 60 /60 | 0;
	var timeMinutes = (time - timeHours*60*60) /60 | 0;
	var timeSeconds = time - timeHours*60*60 - timeMinutes*60;
	var output = "";
	function formatTimeNum(t,l) {
		var output = t + " " + l + " ";					
			if (t > 0) { 
				if (t < 10) {output = "0" + output;} 
			} else { 
				if (l == "s") {
					output = "00 s"
				} else {
					output = ""
				}
			}
		return output
	}				
	return formatTimeNum(timeHours,"h") + formatTimeNum(timeMinutes,"m") + formatTimeNum(timeSeconds,"s")
}

// Generate
var AARFileDetailsBase = function() {
	this.name 		= aarData.metadata.name;
	this.island 	= aarData.metadata.island;
	this.date 		= aarData.metadata.date;
	this.summary	= aarData.metadata.desc;
	this.filename 	= "";
	this.configLine = "";

    this.setFilename = function() {
    	this.filename = "AAR." + this.date + "." + this.island + "." + normalizeFilename(this.name); // this.name.replace(/ /g, '_');
    	this.draw();
    };

    this.updateAAR = function() {
    	this.draw();

        aarData.metadata.name   = $( "#mission-name" ).val();
		aarData.metadata.desc   = $( "#mission-desc" ).val();
		aarData.metadata.island = $( "#mission-island" ).val();
		aarData.metadata.date   = $( "#mission-date" ).val();

        this.name    = aarData.metadata.name;
        this.summary = aarData.metadata.desc;
        this.island  = aarData.metadata.island;
        this.date    = aarData.metadata.date;

        console.log( "AAR text re-generated.");
    };
    this.generateConfigLine = function() {
    	const br = "\n";
    	this.configLine = '	{' +  br + '		"date": "' + this.date
    	+ '"' +  br + '		,"title": "' + this.name
    	+ '"' +  br + '		,"terrain": "' + this.island
    	+ '"' +  br + '		,"link": "aars/' + this.filename + (zipAARFile ? ".zip" : ".txt")
    	+ '"' +  br + '	},';
    }

	this.initEvents = function() {
		$( "#mission-name" ).on('blur', function () {
			AARFileDetails.name = $( "#mission-name" ).val();
			AARFileDetails.setFilename();
		});
		$( "#mission-island" ).on('blur', function () {
			AARFileDetails.island = $( "#mission-island" ).val();
			AARFileDetails.setFilename();
		});
		$( "#mission-date" ).on('blur', function () {
			AARFileDetails.date = $( "#mission-date" ).val();
			AARFileDetails.setFilename();
		});
	};
	this.draw = function() {
		$( "#mission-name" ).val( this.name );
        $( "#mission-filename" ).val( this.filename );
        $( "#mission-island" ).val( this.island );
        $( "#mission-desc" ).val( this.summary );
        $( "#mission-date" ).val( this.date );

		$( "#header-back-btn").css("display", "");
		$('#output-config-line textarea').remove();
		$('#output-config-uploader label').remove();
	};
	this.init = function() {
		this.setFilename();
		this.draw();
		this.initEvents();
		this.configLine = "";
	};

	this.init();
};

function saveGeneratedAARData() {
	AARFileDetails.updateAAR();
	AARFileDetails.generateConfigLine();

	toggleProgressView(true);
	updateProgressView("Saving", "Wait until AAR file will be generated.")

	$('#output-config-uploader').append(
    	"<label id='config-file-btn' class='config-file-btn' for='configUploader'>Update Config File</label>"
    );
	$('#output-config-line').append("<textarea cols='40' rows='6'>" + AARFileDetails.configLine + "</textarea>"	);

	setTimeout ( saveAARFile, 500, "aarFileData = " + JSON.stringify( aarData ) )
}

function saveAARFile(data) {	
	let filename = AARFileDetails.filename;
    let blob = new Blob([data], {type: "text/plain"});	
	if (zipAARFile) {
		saveZipFile(filename, blob);
	} else {
		saveFile(filename + ".txt", blob);
	}
}

function saveZipFile(filename, blob) {
	let onClose = function (blob) { 
		saveFile(filename + ".zip", blob)
	};
	
	zip.createWriter(
		new zip.BlobWriter("application/zip")
		, function (zipWriter) {
			zipWriter.add(filename + ".txt", new zip.BlobReader(blob), function () {
				zipWriter.close(onClose);
			});
		}
		, function (message) { console.error(message) }
	);
}

function saveFile(filename, blob) {
	var a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = filename;
    a.click();
	
	toggleProgressView(false);
};

function saveConfigFile(data) {
	var a = document.createElement("a");
    var file = new Blob([data], {type: "text/plain"});
    a.href = URL.createObjectURL(file);
    a.download = "aarListConfig.ini";
    a.click();
}

/*
function initToggleInterpolate () {
	$('#header-back-btn').on("click", function () {
		if (enableInterpolation) {
			$(this).removeClass("header-btn-selected");
			//enableInterpolation = false;
		} else {
			$(this).addClass("header-btn-selected");
			//enableInterpolation = true;
		};
	});
};
*/
$( document ).ready(function() {
	$( ".dl-2 > input, textarea, button" ).attr( "disabled", "true" );
	resetForm();
	
	zip.workerScriptsPath = "src/js/";
	$('#header-back-btn').on("click", (event)=>{
		showAARParseSuggestion()
	})

	// initToggleInterpolate();
});