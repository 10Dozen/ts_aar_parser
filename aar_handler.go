package main

type AARCoreMeta struct {
	name    string
	terrain string
	guid    string
	summary string
}

type AARUnitMeta struct {
	id       int
	Name     string
	Side     string
	IsPlayer bool
}

type AARVehicleMeta struct {
	id   int
	Name string
}

type ARRUnitFrame struct {
	x       int
	y       int
	dir     int
	IsAlive bool
	// ?
}

type ARRVehicleFrame struct {
	x         int
	y         int
	dir       int
	IsAlive   bool
	Owner     int
	CrewCount int
}

/*

20:15:53 "<AAR-dingor82583><0><unit>[0,8329,2359,168,1,-1]</unit></0></AAR-dingor82583>"
21:04:09 "<AAR-cup_chernarus_A334430><meta><veh>{ ""vehMeta"": [517,""HEMTT Ammo""] }</veh></meta></AAR-cup_chernarus_A334430>"
21:04:13 "<AAR-cup_chernarus_A334430><7><veh>[500,3246,8407,317,1,56,-1]</veh></7></AAR-cup_chernarus_A334430>"

20:15:53 "<AAR-dingor82583><meta><core>{ ""island"": ""dingor"", ""name"": ""CO16 Western"", ""guid"": ""dingor82583"",
""summary"": ""Ковбои освобождают свой городок от бандитов"" }</core></meta></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [0,""Osamich"",""blufor"",1] }</unit></meta></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [1,""Ka6aH"",""blufor"",1] }</unit></meta></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [2,""invaderok"",""blufor"",1] }</unit></meta></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [3,""Smoker"",""blufor"",1] }</unit></meta></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [4,""Реневал"",""blufor"",1] }</unit></meta></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [5,""10Dozen"",""blufor"",1] }</unit></meta></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [6,""chek1"",""blufor"",1] }</unit></meta></AAR-dingor82583>"

20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [16,"""",""civ"",0] }</unit></meta></AAR-dingor82583>"

20:15:53 "<AAR-dingor82583><0><unit>[0,8329,2359,168,1,-1]</unit></0></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><0><unit>[1,8326,2360,168,1,-1]</unit></0></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><0><unit>[2,8324,2359,168,1,-1]</unit></0></AAR-dingor82583>"

20:15:59 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [17,"""",""opfor"",0] }</unit></meta></AAR-dingor82583>"



21:04:13 "<AAR-cup_chernarus_A334430><7><veh>[500,3246,8407,317,1,56,-1]</veh></7></AAR-cup_chernarus_A334430>"
21:04:13 "<AAR-cup_chernarus_A334430><7><veh>[501,2301,9549,313,1,-1,-1]</veh></7></AAR-cup_chernarus_A334430>"
21:04:13 "<AAR-cup_chernarus_A334430><7><veh>[502,3141,8417,271,1,53,-1]</veh></7></AAR-cup_chernarus_A334430>"

21:04:09 "<AAR-cup_chernarus_A334430><meta><veh>{ ""vehMeta"": [517,""HEMTT Ammo""] }</veh></meta></AAR-cup_chernarus_A334430>"
*/
