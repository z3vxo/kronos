package server

import "math/rand"

var codeNameAdjectives = []string{
	"golden", "cursed", "divine", "fallen", "eternal", "sacred", "forsaken", "ancient",
	"wrathful", "fated", "bound", "sunken", "hollow", "immortal", "exiled", "doomed",
	"blessed", "wretched", "proud", "vengeful", "defiant", "lost", "undying", "scorned",
	"fearless", "mighty", "swift", "cunning", "ruthless", "fierce", "wise", "silent",
	"restless", "wandering", "forgotten", "shattered", "risen", "burning", "cold", "blind",
}

var codeNameNouns = []string{
	"titan", "cyclops", "hydra", "chimera", "medusa", "minotaur", "centaur", "harpy",
	"phoenix", "griffin", "cerberus", "gorgon", "sphinx", "siren", "satyr", "nymph",
	"hero", "oracle", "shade", "fury", "muse", "fate", "herald", "specter",
	"argonaut", "spartan", "olympian", "labyrinth", "icarus", "daedalus", "achilles", "odysseus",
	"prometheus", "sisyphus", "tantalus", "orpheus", "heracles", "theseus", "perseus", "jason",
}

func GenCodeName() string {
	adj := codeNameAdjectives[rand.Intn(len(codeNameAdjectives))]
	noun := codeNameNouns[rand.Intn(len(codeNameNouns))]
	return adj + "-" + noun
}
