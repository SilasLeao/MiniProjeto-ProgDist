package remotelist

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

// Estrutura da lista
type RemoteList struct {
	Mu    sync.Mutex
	Lists map[int][]int
	Log   *os.File
}

// Estrutura usada para gravar operações do log
type LogEntry struct {
	Op     string `json:"op"` // operação
	ListID int    `json:"list_id"`
	Value  int    `json:"value"`
}

// Construtor da lista
func NewRemoteList() *RemoteList {
	rl := &RemoteList{
		Lists: make(map[int][]int),
	}
	// Operações de carregar o log e snapshot para manter consistência dos dados assim que a lista é criada
	rl.loadSnapshot()
	rl.loadLog()

	f, _ := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	rl.Log = f

	// inicia goroutine de snapshot para executar em background a cada 10s
	go rl.snapshotRoutine()

	return rl
}

// Métodos RPC

func (rl *RemoteList) Append(args [2]int, reply *bool) error {
	listID := args[0]
	value := args[1]

	rl.Mu.Lock()
	defer rl.Mu.Unlock()

	rl.Lists[listID] = append(rl.Lists[listID], value)
	*reply = true

	rl.writeLog("append", listID, value)
	fmt.Println("Append:", rl.Lists)

	return nil
}

func (rl *RemoteList) Remove(listID int, reply *int) error {
	rl.Mu.Lock()
	defer rl.Mu.Unlock()

	list, ok := rl.Lists[listID]
	if !ok || len(list) == 0 {
		return errors.New("lista vazia")
	}

	val := list[len(list)-1]
	*reply = val
	rl.Lists[listID] = list[:len(list)-1]

	rl.writeLog("remove", listID, val)
	fmt.Println("Remove:", rl.Lists)

	return nil
}

func (rl *RemoteList) Get(args [2]int, reply *int) error {
	listID := args[0]
	pos := args[1]

	rl.Mu.Lock()
	defer rl.Mu.Unlock()

	list, ok := rl.Lists[listID]
	if !ok || pos < 0 || pos >= len(list) {
		return errors.New("índice inválido")
	}

	*reply = list[pos]
	return nil
}

func (rl *RemoteList) Size(listID int, reply *int) error {
	rl.Mu.Lock()
	defer rl.Mu.Unlock()

	*reply = len(rl.Lists[listID])
	return nil
}

// Funções de Log e Snapshot para persistência

func (rl *RemoteList) writeLog(op string, listID, value int) {
	entry := LogEntry{op, listID, value}
	// Converte a entrada do log em JSON
	data, _ := json.Marshal(entry)
	rl.Log.Write(append(data, '\n'))
}

func (rl *RemoteList) loadLog() {
	f, err := os.Open("log.txt")
	if err != nil {
		return
	}
	defer f.Close()

	// Cria um json decoder e entra em um loop para ler o arquivo de log linha a linha, terminando quando encontrar algum erro no caminho, por exemplo linha vazia(JSON inválido)
	decoder := json.NewDecoder(f)
	for {
		var entry LogEntry
		err := decoder.Decode(&entry)
		if err != nil {
			break
		}
		// Refaz as operações de append que achar no log
		if entry.Op == "append" {
			rl.Lists[entry.ListID] = append(rl.Lists[entry.ListID], entry.Value)
		}
		// Refaz as operações de remove que achar no log
		if entry.Op == "remove" {
			list := rl.Lists[entry.ListID]
			if len(list) > 0 {
				rl.Lists[entry.ListID] = list[:len(list)-1]
			}
		}
	}
}

// Rotina de snapshot que roda em background a cada 10s para ficar salvando
func (rl *RemoteList) snapshotRoutine() {
	for {
		time.Sleep(10 * time.Second)
		rl.saveSnapshot()
	}
}

func (rl *RemoteList) saveSnapshot() {
	rl.Mu.Lock()
	defer rl.Mu.Unlock()

	f, _ := os.Create("snapshot.json")
	json.NewEncoder(f).Encode(rl.Lists)
	f.Close()

	// Limpa o log para não ficar enorme com informações redundantes que já foram salvas no snapshot
	rl.Log, _ = os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
}

func (rl *RemoteList) loadSnapshot() {
	f, err := os.Open("snapshot.json")
	if err != nil {
		return
	}
	defer f.Close()

	json.NewDecoder(f).Decode(&rl.Lists)
}
