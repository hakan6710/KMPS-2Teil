package main

import (
	"math/rand"
	"strconv"
	"sync"
	"time"
)

//Konstantendefinition
const Aufzuganzahl=4
const Maximalkapazitaet_Aufzug=5
const Max_Personen=100
const Etagen=10


var wg sync.WaitGroup
var wgEnd sync.WaitGroup
var allePersonenAngekommen=false


var statistikPersonen=make(chan statistikPerson,Max_Personen)
var statistikAufzug=make(chan int,Aufzuganzahl)
type person struct {
	name string

	zieletage int
	startetage int

	startzeit	time.Time
	wartezeit int

	antwortchannel chan int
	eingestiegenbeiWegstrecke int
	gesamtWegstrecke int
}
type aufzug struct{
	name string
	derzeitigesStockwerk int
	gesamtWegstrecke int
	fahrtrichtungNachOben bool
	tickerChan chan int
}

type statistikPerson struct{
	wartezeit int64
	gesamtwegstrecke int
}

type statistikGesamteSimulation struct{
	pStrecke int
	pZeit int64
	aStrecke int
}
func createAufzug(name string)(a aufzug){
	a.name=name
	a.derzeitigesStockwerk=0
	a.gesamtWegstrecke=0
	a.fahrtrichtungNachOben=true
	a.tickerChan=make(chan int,1)
	return a
}

func createPerson(name string) (p person) {
	p.name=name
	p.gesamtWegstrecke=0

	zahlenungleich:=false
	for zahlenungleich==false{
		p.startetage=rand.Intn(Etagen)
		p.zieletage=rand.Intn(Etagen)
		if(p.startetage!=p.zieletage){
			zahlenungleich=true
		}
	}

	p.antwortchannel=make(chan int)

	return p
}

func randomPersonCreat(anfrage_channel chan person){
	//Personenroutine erstellen
	for i:=0; i<Max_Personen;i++{
		personen_name:="P"+strconv.Itoa(i)
		p:=createPerson(personen_name)
		go personen_routine(p,anfrage_channel)
		if(i%2==0){
			time.Sleep(100)
		}
	}
}
func steuerlogik(retChan chan statistikGesamteSimulation){
	//Aufzüge erstellen
	var aufzugChans[Aufzuganzahl]  chan person
	var aufzugTicker[Aufzuganzahl] chan int
	for i:=0;i<Aufzuganzahl;i++{
		aufzugChans[i]=make(chan person)
		aufzug_name:="A"+strconv.Itoa(i)
		a:=createAufzug(aufzug_name)
		aufzugTicker[i]=make(chan int,1)
		a.tickerChan=aufzugTicker[i]
		go aufzug_routine(a,aufzugChans[i])
	}

	//Channel für Anfragen von Personen
	println("Person")
	anfrage_channel:= make(chan person,100)
	go randomPersonCreat(anfrage_channel)
	println("Personerstellt")
	//Aufzugssteuerung aufrufen und auf deren Terminierung warten
	tchan:=make(chan int,1)
	println("X")
	go aufzugsteuerung2(anfrage_channel,aufzugChans,tchan)

	println("Test")
	for allePersonenAngekommen!=true{
		//time.Sleep(10 * time.Millisecond)
		for i:=0;i<Aufzuganzahl;i++{
			print("I=",i)
			aufzugTicker[i]<-0
		}
		println("BLA BLA ",allePersonenAngekommen)
	}
	println("BASDBASDAJSDASDJASDJNASDN")
	var aStatistik=<-tchan

	statistikGesammelt:=false
	WegstreckeAllerPersonen:=0
	WartezeitAllerPersonen:=int64(0)
	counter:=0

	//Personenstatistiken sammeln
	for statistikGesammelt!=true {
		select{
		case msg1:=<-statistikPersonen:
			WegstreckeAllerPersonen+=msg1.gesamtwegstrecke
			WartezeitAllerPersonen+=msg1.wartezeit
			counter+=1
		default:
			if(counter>=Max_Personen){
				statistikGesammelt=true

			}
		}
	}

	retChan<-statistikGesamteSimulation{WegstreckeAllerPersonen,WartezeitAllerPersonen,aStatistik}

	}
func zentrale_steuerlogik() {

	//random generator initialisieren
	rand.NewSource(time.Now().UnixNano())
	retChan:=make(chan statistikGesamteSimulation,1)
	for i:=0;i<4;i++{
		allePersonenAngekommen=false
		go steuerlogik(retChan)
		wg.Add(Max_Personen)
		wg.Wait()
		allePersonenAngekommen=true

		var s1 =<-retChan
		println(s1.pStrecke,"/",s1.pZeit/1000000,"/",s1.aStrecke)
	}

	wgEnd.Done()
}

func aufzugsteuerung1(anfragePersonen chan person,aufzugChans[Aufzuganzahl] chan person ,tchan chan int){
	roundRobinZaehler:=0
	for allePersonenAngekommen!=true {

		ankommendeAnfrage:=<-anfragePersonen
		aufzugChans[roundRobinZaehler]<-ankommendeAnfrage

		if(roundRobinZaehler<Aufzuganzahl-1){
			roundRobinZaehler+=1
		}else {
			roundRobinZaehler=0
		}
		//println(ankommendeAnfrage.name,ankommendeAnfrage.etage)
	}
	tchan<-0
}


func aufzugsteuerung2(anfragePersonen chan person,aufzugChans[Aufzuganzahl] chan person ,tchan chan int){

	var letztePersonZiel[Aufzuganzahl] int

	for i:=0;i<Aufzuganzahl;i++{
		letztePersonZiel[i]=-1
	}
	for allePersonenAngekommen!=true {

		ankommendeAnfrage:=<-anfragePersonen

		//Suche geringste Differenz zwischen Zieletage der Anfrage und dem Wert aus letztePersonZiel
		auswahl:=0
		diffrenz_temp:=0
		for i:=0;i<Aufzuganzahl;i++{
			if(ankommendeAnfrage.zieletage-letztePersonZiel[i]<diffrenz_temp){
				auswahl=i
			}
		}

		aufzugChans[auswahl]<-ankommendeAnfrage

		//println(ankommendeAnfrage.name,ankommendeAnfrage.etage)
	}

	ergebnis:=0
	counter :=0

	//println("Simulation fertig")
	for counter<Aufzuganzahl{
		var msg1=<-statistikAufzug
		ergebnis+=msg1
		counter+=1
	}
	//println("Aufzugstatistik sammeln fertig")
	tchan<-ergebnis
}

func updateStockwerk(a *aufzug){


	if a.fahrtrichtungNachOben{
		if a.derzeitigesStockwerk<Etagen{
			a.derzeitigesStockwerk+=1
			a.gesamtWegstrecke+=1
			}else{
			a.fahrtrichtungNachOben=false
		}
	}else {
		if a.derzeitigesStockwerk>0{
			a.derzeitigesStockwerk-=1
			a.gesamtWegstrecke+=1
		}else{
			a.fahrtrichtungNachOben=true
		}
	}
}
func EinsteigenUndAussteigen(a *aufzug,anfragen*[]person,eingestiegene*[]person){
	//Einsteigenlassen
	for i:=0;i< len(*anfragen);i++{
		if (*anfragen)[i].startetage==a.derzeitigesStockwerk{
			//Anfrage aus der Liste rausloeschen
			(*anfragen)[i].eingestiegenbeiWegstrecke=a.gesamtWegstrecke
			*eingestiegene = append(*eingestiegene,(*anfragen)[i])

			//Anfrage aus der Liste rausloeschen
			(*anfragen) = append((*anfragen)[:i], (*anfragen)[i+1:]...)

		}
	}

	//Aussteigenlassen
	for i:=0;i< len(*eingestiegene);i++{
		if (*eingestiegene)[i].zieletage==a.derzeitigesStockwerk{
			//Antworte Person, die aussteigt mit der gefahrenen Wegstrecke
			(*eingestiegene)[i].antwortchannel<-(a.gesamtWegstrecke-(*eingestiegene)[i].eingestiegenbeiWegstrecke)

			//Anfrage aus der Liste rausloeschen
			(*eingestiegene) = append((*eingestiegene)[:i], (*eingestiegene)[i+1:]...)
		}
	}

}
func aufzug_routine(a aufzug,channel chan person) {

	anfragenListe := make([]person, 0)
	eingestiegenenListe:=make([]person,0)


	for allePersonenAngekommen!=true {
		select{
		case msg1:=<-channel:
			//Füge Anfrage der Person der anfragenListe hinzu
			anfragenListe =append(anfragenListe,msg1)
		default:
		}
		if(len(anfragenListe)+len(eingestiegenenListe)>0){
			updateStockwerk(&a)
			EinsteigenUndAussteigen(&a,&anfragenListe,&eingestiegenenListe)
		}
		println("Vor Ticker")
		<-a.tickerChan
		println("Nach Ticker")
	}

	statistikAufzug<-a.gesamtWegstrecke

}
func personen_routine(p person, anfrageChan chan person) {
	//println("Person erstellt")

	start:=time.Now()

	anfrageChan<-p
	p.gesamtWegstrecke=<-p.antwortchannel
	gemesseneZeit:=time.Since(start)
	wg.Done()
	statistikPersonen<-statistikPerson{gemesseneZeit.Nanoseconds(),p.gesamtWegstrecke}

}

func main() {
	go zentrale_steuerlogik()

	wgEnd.Add(1)
	wgEnd.Wait()
}