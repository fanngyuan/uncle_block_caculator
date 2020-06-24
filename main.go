package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

const mbtc =1000

const totalCOIN=21000000*mbtc
const diffAdjTimeTarget = time.Hour*24*14
const initPerBlockTime = time.Second*60
const uncleBlockRateTarget = 8
const reduceHalfTime = time.Hour*24*365*4
var perAdjCoin = int(totalCOIN/(2*(reduceHalfTime.Seconds()/diffAdjTimeTarget.Seconds())))

// Block represents each 'item' in the blockchain
type Block struct {
	Index      int
	Timestamp  time.Time
	Uncles []string
	Difficulty int
	Nonce uint64
	coin uint64
}

var blocks []Block

func main() {

	genesis := Block{
		Index:      0,
		Timestamp:  time.Now(),
		Difficulty: 1,
		Nonce:0,
		coin:50*mbtc,
	}

	blocks=make([]Block,0,10)
	blocks=append(blocks,genesis)

	makeBlock(genesis)

}

func makeBlock(genesis Block){
	block := genesis
	duration:=initPerBlockTime
	blocksCount:=int(diffAdjTimeTarget.Seconds()/duration.Seconds())-1

	halfCoin:=totalCOIN/2
	perblockCoin:=perAdjCoin/blocksCount

	uncle_func := func ()bool{
		boundry:=rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(uncleBlockRateTarget*2)
		has_uncle := rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(100)<boundry
		return has_uncle
	}
	leftCoin:=  totalCOIN

	for j:=0;j<math.MaxInt32;j++{
		for i:=0;i<=blocksCount;i++{
			block=generateBlock(block,duration,uncle_func)
			blocks=append(blocks,block)
			leftCoin-=perblockCoin
		}
		if leftCoin<halfCoin {
			fmt.Println("开始减半")
			halfCoin = halfCoin/2
			perAdjCoin = perAdjCoin/2
		}
		if j==0{
			duration,blocksCount,uncle_func,perblockCoin =adjustDiff(blocksCount+1,perAdjCoin)
		}else {
			duration,blocksCount,uncle_func,perblockCoin =adjustDiff(blocksCount,perAdjCoin)
		}
		if j==1040{
			break
		}
	}
}

func adjustDiff(blocksCount int,perAdjCoin int)(time.Duration,int,func()bool,int){
	length := len(blocks)
	first:= blocks[length-blocksCount-1]
	last:= blocks[length-1]

	duration:=last.Timestamp.Sub(first.Timestamp)
	// 计算下个周期多久
	nextDurationTarget:=diffAdjTimeTarget.Seconds()-duration.Seconds()+diffAdjTimeTarget.Seconds()
	prevPerblockTime:= int(duration.Seconds())/blocksCount // 上个周期，平均每个块的时间

	// 统计叔块率
	uncles:=0
	for i:=0;i<blocksCount;i++{
		if blocks[length-1-i].Uncles != nil {
			uncles++
		}
	}
	unclesRate:= (float32(uncles)/float32(blocksCount))*100

	// 下个周期出块时间
	nextPerBlockTime:=float64(unclesRate)/float64(uncleBlockRateTarget)*float64(prevPerblockTime)
	nextBlockCount := nextDurationTarget/nextPerBlockTime
	nextPerBlockTimeDuration:=time.Duration(int(nextPerBlockTime/time.Second.Seconds()))*time.Second

	perBlockCoin := perAdjCoin/int(nextBlockCount)

	fmt.Printf("当前叔块率%v，出块时间%v，块数%d ,总时间%v ,coinbase  %v \r\n" ,unclesRate,nextPerBlockTimeDuration,int(nextBlockCount),time.Duration(nextDurationTarget)*time.Second,perBlockCoin)

	// 叔块函数
	uncle_func:=func ()bool{
		data:=int64(0)
		if unclesRate>uncleBlockRateTarget {
			data=-2
		}else {
			data=3
		}
		has_uncle := rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(100)<uncleBlockRateTarget+data
		return has_uncle
	}

	return nextPerBlockTimeDuration,int(nextBlockCount),uncle_func,perBlockCoin
}

func generateBlock(oldBlock Block,duration time.Duration,uncle_func func()bool) Block {
	var newBlock Block

	data:=int64(0)
	for {
		data = rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(int64(duration.Seconds() * 2))
		if data!=0 {
			break
		}
	}

	has_uncle := uncle_func()
	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = time.Unix(oldBlock.Timestamp.Unix()+data,0)
	newBlock.Difficulty = 1
	if has_uncle{
		newBlock.Uncles=[]string {"1"}
	}

	//fmt.Printf("block is %v \r\n",newBlock)

	return newBlock
}