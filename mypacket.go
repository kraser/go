package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/opesun/goquery"

	//Пакеты, которые пригодятся для работы с файлами и сигналами:
	"io"
	"os"
	"os/signal"

	//А вот эти - для высчитывания хешей:
	"crypto/md5"
	"encoding/hex"
)

const (
	FLOWS         int    = 5            //кол-во "потоков"
	REPORT_PERIOD int    = 10           //частота отчетов (сек)
	REP_TO_STOP   int    = 500          //максимум повторов до останова
	HASH_FILE     string = "hash.bin"   //файл с хешами
	QUOTES_FILE   string = "quotes.txt" //файл с цитатами
)

var (
	flows         int
	reportPeriod  int
	repeatsToStop int
	hashFile      string
	quotesFile    string
	used          map[string]bool = make(map[string]bool) //map в котором в качестве ключей будем использовать строки, а для значений - булев тип.
)

func grab() <-chan string { //функция вернет канал, из которого мы будем читать данные типа string
	c := make(chan string)
	for i := 0; i < flows; i++ { //в цикле создадим нужное нам количество гоурутин - worker'oв
		go func() {
			for { //в вечном цикле собираем данные
				x, err := goquery.ParseUrl("http://vpustotu.ru/moderation/")

				if err == nil {
					//fmt.Println(x.HtmlAll())
					//fmt.Println(x.Find(".fi_text").Text())
					//fmt.Println(x.Find(".fi_text").HtmlAll())
					if s := strings.TrimSpace(x.Find(".fi_text").Text()); s != "" {
						c <- s //и отправляем их в канал
					}
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()
	}
	fmt.Println("Запущено потоков: ", flows)
	return c
}

func init() {
	//Задаем правила разбора:
	flag.IntVar(&flows, "w", FLOWS, "количество потоков")
	flag.IntVar(&reportPeriod, "r", REPORT_PERIOD, "частота отчетов (сек)")
	flag.IntVar(&repeatsToStop, "d", REP_TO_STOP, "кол-во дубликатов для остановки")
	flag.StringVar(&hashFile, "hf", HASH_FILE, "файл хешей")
	flag.StringVar(&quotesFile, "qf", QUOTES_FILE, "файл записей")

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	//И запускаем разбор аргументов
	flag.Parse()
	/*
		x, err := goquery.ParseUrl("http://vpustotu.ru/moderation/")
		if err == nil {

			fmt.Println(x.HtmlAll())
			fmt.Println(x.Find(".fi_text").Text())
		}
	*/
	quotes_file, err := os.OpenFile(quotesFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	check(err)
	defer quotes_file.Close()

	//...и файл с хешами
	hash_file, err := os.OpenFile(hashFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	check(err)
	defer hash_file.Close()

	testFileHandler, err := os.OpenFile("my_test.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	check(err)
	defer testFileHandler.Close()

	length, err := testFileHandler.WriteString("Переспала со знакомым просто так и влюбилась. Он несвободен, человек совершенно другого круга и бэкграунда, нищий без профессии")
	check(err)
	fmt.Println(length)

	//Создаем Ticker который будет оповещать нас когда пора отчитываться о работе
	ticker := time.NewTicker(time.Duration(reportPeriod) * time.Second)
	defer ticker.Stop()

	//Создаем канал, который будет ловить сигнал завершения, и привязываем к нему нотификатор...
	key_chan := make(chan os.Signal, 1)
	signal.Notify(key_chan, os.Interrupt)

	//...и все что нужно для подсчета хешей
	hasher := md5.New()

	//Счетчики цитат и дубликатов
	quotes_count, dup_count := 0, 0

	//Все готово, поехали!
	quotes_chan := grab()
	for {
		select {
		case quote := <-quotes_chan: //если "пришла" новая цитата:
			quotes_count++
			//считаем хеш, и конвертируем его в строку:
			hasher.Reset()
			fmt.Println(quote)
			io.WriteString(hasher, quote)
			hash := hasher.Sum(nil)
			hash_string := hex.EncodeToString(hash)
			//проверяем уникальность хеша цитаты
			if !used[hash_string] {
				//все в порядке - заносим хеш в хранилище, и записываем его и цитату в файлы
				used[hash_string] = true
				hash_file.Write(hash)
				quotes_file.WriteString(quote + "\n\n\n")
				dup_count = 0
			} else {
				//получен повтор - пришло время проверить, не пора ли закругляться?
				if dup_count++; dup_count == repeatsToStop {
					fmt.Println("Достигнут предел повторов, завершаю работу. Всего записей: ", len(used))
					return
				}
			}
		case <-key_chan: //если пришла информация от нотификатора сигналов:
			fmt.Println("CTRL-C: Завершаю работу. Всего записей: ", len(used))
			return
		case <-ticker.C: //и, наконец, проверяем не пора ли вывести очередной отчет
			fmt.Printf("Всего %d / Повторов %d (%d записей/сек) \n", len(used), dup_count, quotes_count/reportPeriod)
			quotes_count = 0
		}
	}

}
