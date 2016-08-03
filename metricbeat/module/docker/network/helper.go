package network

import (
	"time"
	dc "github.com/fsouza/go-dockerclient"
	"github.com/elastic/beats/metricbeat/module/docker"
	"github.com/elastic/beats/libbeat/logp"
	"fmt"
)

type NETService struct {
	NetworkStatPerContainer map[string]map[string]NETRaw // IDCOntainer : interfaceName :
}
type NETRaw struct {
	Time time.Time
	RxBytes   uint64
	RxDropped uint64
	RxErrors  uint64
	RxPackets uint64
	TxBytes   uint64
	TxDropped uint64
	TxErrors  uint64
	TxPackets uint64
}
type NETstats struct {
	Time      time.Time
	MyContainer *docker.Container
	NameInterface string
	RxBytes   float64
	RxDropped float64
	RxErrors  float64
	RxPackets float64
	TxBytes   float64
	TxDropped float64
	TxErrors  float64
	TxPackets float64
}
/*func (NT NETService) getNetworkStatsList ()map[string][]NETstats {

}*/
func (NT NETService) GetNetworkStatsPerContainer( rawStats []docker.DockerStat) []NETstats{
	formatedStats := []NETstats{}
	if len(rawStats) !=0 {
		for _,myStats := range rawStats {
			for nameInterface, rawnNetStats := range myStats.Stats.Networks{
				formatedStats = append(formatedStats, NT.getNetworkStats(nameInterface,myStats.Stats.Read,&myStats.Container,&rawnNetStats))
			}
		}
	}else{
		fmt.Printf("No container is running \n")
	}


	return formatedStats
}
func (NT NETService) checkStats (containerID string, nameInterface string) bool{
	if _, exist := NT.NetworkStatPerContainer[containerID][nameInterface]; exist{
		return true
	}
	return false

}
func (NT NETService) getNetworkStats (  nameInterface string, time time.Time, container *dc.APIContainers, rawStats * dc.NetworkStats) NETstats{

	 myNETstats := NETstats{}
	if rawStats != nil{

		newNetworkStats := getNewNet(time,rawStats)
		oldNetworkStat, exist := NT.NetworkStatPerContainer[container.ID][nameInterface]
		//fmt.Printf(" EXIST : ", exist)
		//fmt.Printf(" NEW RXBYTES: ", newNetworkStats.RxBytes)
		//fmt.Printf(" OLD RXBYTES:", oldNetworkStat.RxBytes)
		if(exist) {
			myNETstats = NETstats{
				MyContainer: docker.InitCurrentContainer(container),
				Time: time,
				NameInterface: nameInterface,
				RxBytes: NT.getRxBytesPerSecond(&newNetworkStats,&oldNetworkStat),
				RxDropped: NT.getRxDroppedPerSecond(&newNetworkStats,&oldNetworkStat),
				RxErrors: NT.getRxErrorsPerSecond(&newNetworkStats, &oldNetworkStat),
				RxPackets: NT.getRxPacketsPerSecond(&newNetworkStats,&oldNetworkStat),
				TxBytes: NT.getTxBytesPerSecond(&newNetworkStats, &oldNetworkStat),
				TxDropped: NT.getTxDroppedPerSecond(&newNetworkStats,&oldNetworkStat),
				TxErrors: NT.getTxErrorsPerSecond(&newNetworkStats, &oldNetworkStat),
				TxPackets: NT.getTxPacketsPerSecond(&newNetworkStats,&oldNetworkStat),
			}
		}else {
			myNETstats = NETstats{
				MyContainer: docker.InitCurrentContainer(container),
				Time: time,
				NameInterface: nameInterface,
				RxBytes: 0,
				RxDropped:0,
				RxErrors:  0,
				RxPackets: 0,
				TxBytes:   0,
				TxDropped: 0,
				TxErrors:  0,
				TxPackets: 0,
			}
		}
		 if _,exist := NT.NetworkStatPerContainer[container.ID]; !exist{
			NT.NetworkStatPerContainer[container.ID]= make (map [string]NETRaw)
		}
		NT.NetworkStatPerContainer[container.ID][nameInterface] = newNetworkStats


	}else{
		logp.Info(" Unvailable stats ")
	}
	return myNETstats

}

func getNewNet(time  time.Time, stats *dc.NetworkStats) NETRaw{
	return NETRaw{
		Time: time,
		RxBytes: stats.RxBytes,
		RxDropped: stats.RxDropped,
		RxErrors: stats.RxErrors,
		RxPackets: stats.RxPackets,
		TxBytes: stats.TxBytes,
		TxDropped: stats.TxDropped,
		TxErrors: stats.TxErrors,
		TxPackets: stats.TxPackets,

	}


}
func (NT NETService) getRxBytesPerSecond(newStats *NETRaw, oldStats *NETRaw) float64 {
	duration := newStats.Time.Sub(oldStats.Time)
	return NT.calculatePerSecond(duration, oldStats.RxBytes, newStats.RxBytes)
}

func (NT NETService) getRxDroppedPerSecond(newStats *NETRaw, oldStats *NETRaw) float64 {
	duration := newStats.Time.Sub(oldStats.Time)
	return NT.calculatePerSecond(duration, oldStats.RxDropped, newStats.RxDropped)
}

func (NT NETService) getRxErrorsPerSecond(newStats *NETRaw, oldStats *NETRaw) float64 {
	duration := newStats.Time.Sub(oldStats.Time)
	return NT.calculatePerSecond(duration, oldStats.RxErrors, newStats.RxErrors)
}

func (NT NETService) getRxPacketsPerSecond(newStats *NETRaw, oldStats *NETRaw) float64 {
	duration := newStats.Time.Sub(oldStats.Time)
	return NT.calculatePerSecond(duration,oldStats.RxPackets, newStats.RxPackets)
}

func (NT NETService) getTxBytesPerSecond(newStats *NETRaw, oldStats *NETRaw) float64 {
	duration := newStats.Time.Sub(oldStats.Time)
	return NT.calculatePerSecond(duration, oldStats.TxBytes, newStats.TxBytes)
}

func (NT NETService) getTxDroppedPerSecond(newStats *NETRaw, oldStats *NETRaw) float64 {
	duration := newStats.Time.Sub(oldStats.Time)
	return NT.calculatePerSecond(duration, oldStats.TxDropped, newStats.TxDropped)
}

func (NT NETService) getTxErrorsPerSecond(newStats *NETRaw, oldStats *NETRaw) float64 {
	duration := newStats.Time.Sub(oldStats.Time)
	return NT.calculatePerSecond(duration, oldStats.TxErrors, newStats.TxErrors)
}

func (NT NETService) getTxPacketsPerSecond(newStats *NETRaw, oldStats *NETRaw) float64 {
	duration := newStats.Time.Sub(oldStats.Time)
	return NT.calculatePerSecond(duration ,oldStats.TxPackets, newStats.TxPackets)
}

func (NT NETService) calculatePerSecond(duration time.Duration ,oldValue uint64, newValue uint64) float64 {
	//duration := newValue.Time.Sub(oldValue.Time)
	return float64((newValue - oldValue)) / duration.Seconds()
}


