# coinbitlyv2

func main() {
	e := []float64{2699.0, 2680.8, 2650.2}
	r := []float64{2670.8, 2640.2}
	fmt.Printf(" entry1 before reslicing : %v\n", e)
	fmt.Printf(" recover1 before reslicing : %v\n", r)
	fmt.Printf(" entry2 after reslicing : %v\n", e[:len(e)-1])
	fmt.Printf(" recover2 after reslicing : %v\n", r[:len(r)-1])
	fmt.Printf(" entry3 after reslicing : %v\n", e[:len(e)-2])
	fmt.Printf(" recover3 after reslicing : %v\n", r[:len(r)-2])
	h := r[:len(r)-2]
	fmt.Printf("len of h is = %v", len(h))
}








