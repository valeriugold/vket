package price

type Pricer interface {
	PriceGet() (int, error)
}

type MovieDescriptor struct {
	ID	    int
	duration    int
	size	    int
}

type EventPricer interface {
	PriceGet() (int, error)
	PriceSetPerMin(pm int)
	PriceSetPerSize(ps int)
	PriceAddMovie(ID, duration, size int)
	PriceDebugList()
}

type EventPrice struct {
	ID	    int
	Movies	    []MovieDescriptor
	PricePer100M int
	PricePerMinute int
}

func (ep *EventPrice) PriceGet() int {
	td := (int)0
	ts := (int)0
	for _, m := range ep->Movies {
		td += m.duration
		ts += m.size
	}
	p := td * pricePerMinute / 60 + ts * pricePer100M / 100
	return p
}

func (ep *EventPrice) MovieAdd(ID, duration, size int) (int, error) {
	// check if ID already exists
	for _, m := range ep->Movies {
		if ID == m.ID {
			return errors.New("movie " + ID + "already exists")
		}
	}
	Movies = append(Movies, MovieDescriptor{ID, duration, size})
}

func (ep *EventPrice) MovieRemove(ID, duration, size int) error {
	for i, m := range ep->Movies {
		if ID == m.ID {
			// movie exists, delete it
			ep->Movies[i] = a[len(ep->Movies)-1] 
			ep->Movies = ep->Movies[:len(ep->Movies)-1]
			return nill
		}
	}
	return nil
}

type PricerEvent struct {
	AddMovie(duration, resolution, size)
	
}

f() {
	event(movieA, movieB, movieC)
	// to get price
	for movies in event {
		AddMovie
	}
	p := PriceGet()
	AddMovie(100, 200)
}
