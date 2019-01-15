package main

import (
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"
)

//Konstantendefinition
const Aufzuganzahl=4
const Max_Personen=100
const Etagen=10

var wgPersonen sync.WaitGroup
var wgEnd sync.WaitGroup

var allePersonenAngekommen=false

var s1 = rand.NewSource(time.Now().UnixNano())
var r1 = rand.New(s1)

var statistikPersonen=make(chan statistikPerson,Max_Personen)
var statistikAufzug=make(chan int,Aufzuganzahl)

type person struct {
	name string
	nummer int

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
	nummer int
	derzeitigesStockwerk int
	gesamtWegstrecke int
	fahrtrichtungNachOben bool
	tickerChan chan bool
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
func createAufzug(name string,num int)(a aufzug){
	a.name=name
	a.derzeitigesStockwerk=0
	a.gesamtWegstrecke=0
	a.fahrtrichtungNachOben=true
	a.tickerChan=make(chan bool,1)
	a.nummer=num
	return a
}

func createPerson(name string,num int) (p person) {
	p.name=name
	p.gesamtWegstrecke=0
	p.nummer=num
	zahlenidentisch:=true


	for zahlenidentisch{
		p.startetage=r1.Intn(Etagen)
		p.zieletage=r1.Intn(Etagen)
		if(p.startetage!=p.zieletage){
			zahlenidentisch=false
		}
	}
	p.antwortchannel=make(chan int)
	return p
}

func randomPersonCreat(anfrage_channel chan person){

	for i:=0; i<Max_Personen;i++{
		personen_name:="P"+strconv.Itoa(i)
		p:=createPerson(personen_name,i)
		p.startzeit=time.Now()
		go personen_routine(p,anfrage_channel)
		if(i%2==0){
			time.Sleep(time.Duration(r1.Intn(100)))
		}
	}
}

func steuerlogik(retChan chan statistikGesamteSimulation,aufzugsteuerung int){

	var aufzugChans[Aufzuganzahl]  chan person
	var aufzugTicker[Aufzuganzahl] chan bool

	//Aufzug erzeugen
	for i:=0;i<Aufzuganzahl;i++{
		aufzugChans[i]=make(chan person)
		aufzug_name:="A"+strconv.Itoa(i)
		a:=createAufzug(aufzug_name,i)
		aufzugTicker[i]=make(chan bool,1)
		a.tickerChan=aufzugTicker[i]
		go aufzug_routine(a,aufzugChans[i])
	}

	//Channel für Anfragen von Personen
	anfrage_channel:= make(chan person,100)

	//Aufruf des PersonenCreators
	go randomPersonCreat(anfrage_channel)

	//Auswahl von aufzugssteuerungs-algorithmus
	tchan:=make(chan int,1)
	if(aufzugsteuerung==0){
		go aufzugsteuerung1(anfrage_channel,aufzugChans,tchan)
	}else{
		go aufzugsteuerung2(anfrage_channel,aufzugChans,tchan)
	}

	//Ticker für Aufzüge
	var fertig=true
	for fertig{
		for i:=0;i<Aufzuganzahl;i++{
			dontBlock:=true
			for dontBlock{
				select{
				case aufzugTicker[i]<-true:
					dontBlock=false
				default:
					if allePersonenAngekommen==true{
						dontBlock=false
					}
					runtime.Gosched()
				}
			}

			dontBlock=true
			for dontBlock{
				select{
				case <-aufzugTicker[i]:
					dontBlock=false
				default:
					if allePersonenAngekommen==true{
						dontBlock=false
					}
					runtime.Gosched()
				}
			}

		}
		if(allePersonenAngekommen==true){
			fertig=false
		}
	}

	//Auf Statistik von Aufzugssteuerung warten
	var aStatistik=<-tchan

	//Statistik von Personen sammeln
	statistikGesammelt:=true
	WegstreckeAllerPersonen:=0
	WartezeitAllerPersonen:=int64(0)
	counter:=0
	for statistikGesammelt{
		select{
		case msg1:=<-statistikPersonen:
			WegstreckeAllerPersonen+=msg1.gesamtwegstrecke
			WartezeitAllerPersonen+=msg1.wartezeit
			counter+=1
		default:
			if(counter>=Max_Personen){
				statistikGesammelt=false

			}
		}
	}

	//Statistik an Zentrale_Steuerlogik senden
	retChan<-statistikGesamteSimulation{WegstreckeAllerPersonen,WartezeitAllerPersonen,aStatistik}

	}
func zentrale_steuerlogik() {

	retChan:=make(chan statistikGesamteSimulation,1)

	for j:=0;j<2;j++{
		GesamteStatistik :=statistikGesamteSimulation{0,0,0}
		for i:=0;i<3;i++{
			//togglen der Globalen Variable zum Simulationsstopp
			allePersonenAngekommen=false

			//Aufruf der Steuerlogik mit j als Auswahl der Aufzugssteuerungs-Algorithmus
			//retChan als Channel fürs Senden der Stastik
			go steuerlogik(retChan,j)

			//Warte bis alle Personen am Ziel
			wgPersonen.Add(Max_Personen)
			wgPersonen.Wait()

			//Simulationsstopp
			allePersonenAngekommen=true

			//Warte auf Statistik von Steuerlogik
			var s1 =<-retChan
			println("pStrecke=",s1.pStrecke,"pZeit=",s1.pZeit/1000000,"aStrecke=",s1.aStrecke)

			//Statistik des jetzigen Durchlaufs abspeichern
			GesamteStatistik.aStrecke+=s1.aStrecke
			GesamteStatistik.pStrecke+=s1.pStrecke
			GesamteStatistik.pZeit+=s1.pZeit
		}
		println("Aufzugsteuerung",j, ")","Personen_Gesamtstrecke=", GesamteStatistik.pStrecke,"Personen_Zeit=", GesamteStatistik.pZeit,"Aufzug_Gesamtstrecke=", GesamteStatistik.aStrecke)
	}


	wgEnd.Done()
}

func aufzugsteuerung1(anfragePersonen chan person,aufzugChans[Aufzuganzahl] chan person ,tchan chan int){

	//Ankommende Anfrange, nach Roundrobin an Aufzug weitergeben
	roundRobinZaehler:=0
	for allePersonenAngekommen!=true {
		select{
		case msg1:=<-anfragePersonen:
			ankommendeAnfrage:=msg1
			aufzugChans[roundRobinZaehler]<-ankommendeAnfrage

			if(roundRobinZaehler<Aufzuganzahl-1){
				roundRobinZaehler+=1
			}else {
				roundRobinZaehler=0
			}
		default:
		}

	}

	//Statistik sammeln
	ergebnis:=0
	counter :=0
	for counter<Aufzuganzahl{
		var msg1=<-statistikAufzug
		ergebnis+=msg1
		counter+=1
	}
	//Statistik an Steuerlogik senden
	tchan<-ergebnis


}


func aufzugsteuerung2(anfragePersonen chan person,aufzugChans[Aufzuganzahl] chan person ,tchan chan int){

	//Array für Zieletage der zuletzt gesendeten Personanfrage an Aufzug
	var letztePersonZiel[Aufzuganzahl] int
	for i:=0;i<Aufzuganzahl;i++{
		letztePersonZiel[i]=-1
	}

	//Steuerlogik, Suche min(letztePersonziel-anfrage.Zieletage)
	for allePersonenAngekommen!=true {
		select{
		case msg1:=<-anfragePersonen:
			ankommendeAnfrage:=msg1
			//Suche geringste Differenz zwischen Zieletage der Anfrage und dem Wert aus letztePersonZiel
			auswahl:=0
			diffrenz_temp:=0
			for i:=0;i<Aufzuganzahl;i++{
				if(ankommendeAnfrage.zieletage-letztePersonZiel[i]<diffrenz_temp){
					auswahl=i
				}
				if(letztePersonZiel[i]==-1){
					auswahl=i
				}
			}
			aufzugChans[auswahl]<-ankommendeAnfrage
		default:
		}
	}

	//Statistik sammeln von Aufzügen
	ergebnis:=0
	counter :=0
	for counter<Aufzuganzahl{
		var msg1=<-statistikAufzug
		ergebnis+=msg1
		counter+=1
	}

	//Senden der Statistik
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

	//Slices für Personenanfrangen und eingestiegene Personen
	anfragenListe := make([]person, 0)
	eingestiegenenListe:=make([]person,0)


	for allePersonenAngekommen!=true{
		//Ticker für Aufzug
		dontBlock:=true
		for dontBlock{
			select{
			case <-a.tickerChan:
				dontBlock=false
			default:
				if allePersonenAngekommen==true{
					dontBlock=false
				}
				runtime.Gosched()
			}
		}

		//Anfrage von einer Person empfangen
		select{
		case msg1:=<-channel:
			//Füge Anfrage der Person der anfragenListe hinzu
			anfragenListe =append(anfragenListe,msg1)
		default:
		}

		//Abfrage ob überhaupt Personen in den Listen gespeichert
		//je nach dem Stockwerk updaten und Personen einsteigen/aussteigen lassen
		if(len(anfragenListe)+len(eingestiegenenListe)>0){
			updateStockwerk(&a)
			EinsteigenUndAussteigen(&a,&anfragenListe,&eingestiegenenListe)
		}

		//Ticker für Aufzug
		dontBlock=true
		for dontBlock{
			select{
			case a.tickerChan<-true:
				dontBlock=false
			default:
				if allePersonenAngekommen==true{
					dontBlock=false
				}
				runtime.Gosched()
			}
		}

	}

	//Statistik senden
	statistikAufzug<-a.gesamtWegstrecke

}
func personen_routine(p person, anfrageChan chan person) {
	//Anfrage an Aufzugssteuerung senden
	anfrageChan<-p

	//Antwort von Aufzug über Wegstrecke
	p.gesamtWegstrecke=<-p.antwortchannel

	//Berechnung Zeit
	gemesseneZeit:=time.Since(p.startzeit)

	//wg Triggern
	wgPersonen.Done()

	//statistik senden
	statistikPersonen<-statistikPerson{gemesseneZeit.Nanoseconds(),p.gesamtWegstrecke}

}

func main() {
	//Aufruf der zentralen_steuerlogik
	go zentrale_steuerlogik()
	wgEnd.Add(1)
	wgEnd.Wait()
}