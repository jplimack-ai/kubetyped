package reject_wins_over_ignore

// v1/Pod is in both reject_gvks and ignore_gvks — reject wins.
var manifest = map[string]any{ // want `construction of v1/Pod is rejected by project policy \(reject_gvks\)`
	"apiVersion": "v1",
	"kind":       "Pod",
}
