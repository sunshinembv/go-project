package main

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

var (
	TicketAlreadyBookedError = errors.New("Ticket already booked")
	InvalidTicketIDError     = errors.New("invalid ticket id")
)

type BookingTrainTicketsService struct {
	tickets []*Ticket
}

func NewBookingTrainTicketsService() *BookingTrainTicketsService {
	tickets := make([]*Ticket, 100)

	for i := range tickets {
		tickets[i] = NewTicket(i)
	}

	return &BookingTrainTicketsService{
		tickets: tickets,
	}
}

type BookingRequest struct {
	agency   *Agency
	ticketID int
}

func (bs *BookingTrainTicketsService) Book(req BookingRequest) (*Ticket, error) {
	if req.ticketID < 0 || req.ticketID >= len(bs.tickets) {
		return nil, InvalidTicketIDError
	}

	ticket := bs.tickets[req.ticketID]
	if !ticket.isAvailable.CompareAndSwap(true, false) {
		return nil, TicketAlreadyBookedError
	}
	ticket.AgencyName = req.agency.name
	return ticket, nil
}

type Ticket struct {
	ticketID    int
	AgencyName  string
	isAvailable atomic.Bool
}

func NewTicket(id int) *Ticket {
	ticket := &Ticket{
		ticketID: id,
	}
	ticket.isAvailable.Store(true)
	return ticket
}

type Agency struct {
	name          string
	bookedTickets chan *Ticket
}

func NewAgency(id int) *Agency {
	return &Agency{
		name:          fmt.Sprintf("AgencyName_%d", id),
		bookedTickets: make(chan *Ticket),
	}
}

func (a *Agency) GenerateBookingRequest(interrupt chan struct{}) chan BookingRequest {
	req := make(chan BookingRequest)
	go func() {
		defer close(req)
		for i := range 100 {
			select {
			case <-interrupt:
				return
			case req <- BookingRequest{
				agency:   a,
				ticketID: i,
			}:
			}
		}
	}()
	return req
}

func (a *Agency) AgencyBookedTickets(bs *BookingTrainTicketsService, interrupt chan struct{}, bookingReq chan BookingRequest) chan *Ticket {
	go func() {
		defer close(a.bookedTickets)

		for br := range bookingReq {
			bookedTicket, err := bs.Book(br)
			if err != nil {
				continue
			}

			select {
			case <-interrupt:
				return
			case a.bookedTickets <- bookedTicket:
			}
		}
	}()
	return a.bookedTickets
}

func AgencyFanOut(bs *BookingTrainTicketsService, interrupt chan struct{}) []chan *Ticket {
	numAgency := 10
	channels := make([]chan *Ticket, numAgency)
	agencies := make([]*Agency, numAgency)

	for i := range len(agencies) {
		agencies[i] = NewAgency(i)
	}

	for i := range numAgency {
		bookingReq := agencies[i].GenerateBookingRequest(interrupt)
		res := agencies[i].AgencyBookedTickets(bs, interrupt, bookingReq)
		channels[i] = res
	}

	return channels
}

func fanIn(interrupt chan struct{}, resultsChs ...chan *Ticket) chan *Ticket {
	final := make(chan *Ticket)

	var wg sync.WaitGroup

	for _, ch := range resultsChs {
		chCopy := ch

		wg.Go(func() {
			for data := range chCopy {
				select {
				case <-interrupt:
					return
				case final <- data:
				}
			}
		})
	}

	go func() {
		wg.Wait()
		close(final)
	}()

	return final
}

func main() {
	bs := NewBookingTrainTicketsService()

	interrupt := make(chan struct{})
	defer close(interrupt)

	agencyFunOut := AgencyFanOut(bs, interrupt)
	results := fanIn(interrupt, agencyFunOut...)

	count := 0

	for ticket := range results {
		count++
		fmt.Printf("Ticket %d booked by %s\n", ticket.ticketID, ticket.AgencyName)
	}

	fmt.Printf("Total booked tickets: %d\n", count)
}
