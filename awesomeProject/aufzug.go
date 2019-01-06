package main

import (
	"fmt"
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
var defaultPerson=person{"default_person",0,0,make(chan int),0,0}

type person struct {
	name string

	zieletage int
	startetage int

	antwortchannel chan int
	eingestiegenbeiWegstrecke int
	gesamtWegstrecke int
}

type aufzug struct{
	name string
	derzeitigesStockwerk int
	gesamtWegstrecke int
	fahrtrichtungNachOben bool
}


func createAufzug(name string)(a aufzug){
	a.name=name
	a.derzeitigesStockwerk=0
	a.gesamtWegstrecke=0
	a.fahrtrichtungNachOben=true
	return a
}

func createPerson(name string) (p person) {
	p.name=name
	p.gesamtWegstrecke=0
	p.startetage=rand.Intn(Etagen)
	p.zieletage=rand.Intn(Etagen)
	return p
}

func zentrale_steuerlogik() {

	//random generator initialisieren
	rand.NewSource(time.Now().UnixNano())


	//Aufzüge erstellen
	var aufzugChans[Aufzuganzahl]  chan person
	for i:=0;i<Aufzuganzahl;i++{
		aufzugChans[i]=make(chan person)
		aufzug_name:="A"+strconv.Itoa(i)
		a:=createAufzug(aufzug_name)
		go aufzug_routine(a,aufzugChans[i])
	}

	//Channel für Anfragen von Personen
	anfrage_channel:= make(chan person,100)


	//Personenroutine erstellen
	for i:=0; i<Max_Personen;i++{
		personen_name:="P"+strconv.Itoa(i)
		p:=createPerson(personen_name)
		go personen_routine(p,anfrage_channel)
	}
	roundRobinZaehler:=0
	for {

		ankommendeAnfrage:=<-anfrage_channel


		aufzugChans[roundRobinZaehler]<-ankommendeAnfrage

		if(roundRobinZaehler<Aufzuganzahl-1){
			roundRobinZaehler+=1
		}else {
			roundRobinZaehler=0
		}
		//println(ankommendeAnfrage.name,ankommendeAnfrage.etage)
	}
}

func aufzugsteuerung1(){

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
			print("JMD IST EINGETIEGEN")
		}
	}

	//Aussteigenlassen
	for i:=0;i< len(*eingestiegene);i++{
		if (*eingestiegene)[i].zieletage==a.derzeitigesStockwerk{

			//Antworte Person, die aussteigt mit der gefahrenen Wegstrecke
			println("Vor der Kommunikation mit Person")
			(*eingestiegene)[i].antwortchannel<-(a.gesamtWegstrecke-(*eingestiegene)[i].eingestiegenbeiWegstrecke)
			println("Nach der Kommunikation mit Person")
			//Anfrage aus der Liste rausloeschen
			(*eingestiegene) = append((*eingestiegene)[:i], (*eingestiegene)[i+1:]...)
			print("Jmd ist Ausgestiegen")
		}
	}

}
func aufzug_routine(a aufzug,channel chan person) {

	println("Aufzug erstellt")
	anfragenListe := make([]person, 0)
	eingestiegenenListe:=make([]person,0)
	for {
		select{
		case msg1:=<-channel:
			//Füge Anfrage der Person der anfragenListe hinzu
			fmt.Println("Anfrage angekommen")
			anfragenListe =append(anfragenListe,msg1)
		default:
		}

		updateStockwerk(&a)
		EinsteigenUndAussteigen(&a,&anfragenListe,&eingestiegenenListe)
	}

}

func personen_routine(p person, anfrageChan chan person) {

	//print("Person erstellt")
	anfrageChan<-p
	fmt.Println(p.name," ",p.startetage,"/",p.zieletage," ist solange=",strconv.Itoa(<-p.antwortchannel),"mitgefahren")
	wg.Done()
}

func main() {
	go zentrale_steuerlogik()
	wg.Add(Max_Personen)
	wg.Wait()
	}