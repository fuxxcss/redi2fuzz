package fuxx

type Candidate struct {

	graph *Graph
	command string
}

func MutateDumb(corpus *Corpus) string {

	return nil
}

func MutateGramfree(corpus *Corpus) string {

	rand.Seed(time.Now().UnixNano())

	// mutated len
	len := rand.Intn(CORPUS_MAXLEN - CORPUS_MINLEN) + CORPUS_MINLEN
	mutated := ""

	graphSlice := make([]*Graph,1)
	for i := 0 ; i < len ; ++ i {

		// select one command
		testPtr,cmdIndex := corpus.Select()
		graph := testPtr.graph[cmdIndex]

		command := testPtr.commands[cmdIndex][CMD_TEXT]

		// match command
		if len(graph.cmdV.Prev) {

			isMatched := false
			for _,g := range graphSlice {

				// match succeed
				if isMatched = graph.Match(g) ; isMatched {
					break
				}
			}

			// match failed
			if !isMatched {
				continue
			}
		}

		graphSlice = append(graphSlice,graph)
		mutated += graph.cmdV.Data + RediSep
	}

	return mutated

}

func MutateFagent(corpus *Corpus) string {

	return nil
}