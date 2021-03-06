package resolver

import (
	"testing"
	"strings"
	
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	opregistry "github.com/operator-framework/operator-registry/pkg/registry"

	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha2"
)

func buildAPIOperatorGroup(namespace, name string, targets []string, gvks []string) v1alpha2.OperatorGroup {
	return v1alpha2.OperatorGroup {
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name: name,
			Annotations: map[string]string{
				v1alpha2.OperatorGroupProvidedAPIsAnnotationKey: strings.Join(gvks, ","),
			},
		},
		Status: v1alpha2.OperatorGroupStatus{
			Namespaces: targets,
		},
	}
}
func TestNewOperatorGroup(t *testing.T) {
	tests := []struct {
		name string
		in  v1alpha2.OperatorGroup
		want *OperatorGroup
	}{
		{
			name: "NoTargetNamespaces/NoProvidedAPIs",
			in: buildAPIOperatorGroup("ns", "empty-group", nil, nil),
			want: &OperatorGroup{
				namespace: "ns",
				name: "empty-group",
				targets: make(NamespaceSet),
				providedAPIs: make(APISet),
			},
		},
		{
			name: "OneTargetNamespace/NoProvidedAPIs",
			in: buildAPIOperatorGroup("ns", "empty-group", []string{"ns-1"}, nil),
			want: &OperatorGroup{
				namespace: "ns",
				name: "empty-group",
				targets: NamespaceSet{
					"ns": {},
					"ns-1": {},
				},
				providedAPIs: make(APISet),
			},
		},
		{
			name: "OwnTargetNamespace/NoProvidedAPIs",
			in: buildAPIOperatorGroup("ns", "empty-group", []string{"ns"}, nil),
			want: &OperatorGroup{
				namespace: "ns",
				name: "empty-group",
				targets: NamespaceSet{
					"ns": {},
				},
				providedAPIs: make(APISet),
			},
		},
		{
			name: "MultipleTargetNamespaces/NoProvidedAPIs",
			in: buildAPIOperatorGroup("ns", "empty-group", []string{"ns-1", "ns-2"}, nil),
			want: &OperatorGroup{
				namespace: "ns",
				name: "empty-group",
				targets: NamespaceSet{
					"ns": {},
					"ns-1": {},
					"ns-2" :{},
				},
				providedAPIs: make(APISet),
			},
		},
		{
			name: "AllTargetNamespaces/NoProvidedAPIs",
			in: buildAPIOperatorGroup("ns", "empty-group", []string{metav1.NamespaceAll}, nil),
			want: &OperatorGroup{
				namespace: "ns",
				name: "empty-group",
				targets: NamespaceSet{
					metav1.NamespaceAll: {},
				},
				providedAPIs: make(APISet),
			},
		},
		{
			name: "OneTargetNamespace/OneProvidedAPI",
			in: buildAPIOperatorGroup("ns", "group", []string{"ns-1"}, []string{"Goose.v1alpha1.birds.com"}),
			want: &OperatorGroup{
				namespace: "ns",
				name: "group",
				targets: NamespaceSet{
					"ns": {},
					"ns-1": {},
				},
				providedAPIs: APISet{
					opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
				},
			},
		},
		{
			name: "OneTargetNamespace/BadProvidedAPI",
			in: buildAPIOperatorGroup("ns", "group", []string{"ns-1"}, []string{"Goose.v1alpha1"}),
			want: &OperatorGroup{
				namespace: "ns",
				name: "group",
				targets: NamespaceSet{
					"ns": {},
					"ns-1": {},
				},
				providedAPIs: make(APISet),
			},
		},
		{
			name: "OneTargetNamespace/MultipleProvidedAPIs/OneBad",
			in: buildAPIOperatorGroup("ns", "group", []string{"ns-1"}, []string{"Goose.v1alpha1,Moose.v1alpha1.mammals.com"}),
			want: &OperatorGroup{
				namespace: "ns",
				name: "group",
				targets: NamespaceSet{
					"ns": {},
					"ns-1": {},
				},
				providedAPIs: APISet{
					opregistry.APIKey{Group: "mammals.com", Version: "v1alpha1", Kind: "Moose"}: {},
				},
			},
		},
		{
			name: "OneTargetNamespace/MultipleProvidedAPIs",
			in: buildAPIOperatorGroup("ns", "group", []string{"ns-1"}, []string{"Goose.v1alpha1.birds.com,Moose.v1alpha1.mammals.com"}),
			want: &OperatorGroup{
				namespace: "ns",
				name: "group",
				targets: NamespaceSet{
					"ns": {},
					"ns-1": {},
				},
				providedAPIs: APISet{
					opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
					opregistry.APIKey{Group: "mammals.com", Version: "v1alpha1", Kind: "Moose"}: {},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := NewOperatorGroup(tt.in)
			require.NotNil(t, group)
			require.EqualValues(t, tt.want, group)
		})
	}
}

func TestNamespaceSetIntersection(t *testing.T) {
	type input struct {
		left NamespaceSet
		right NamespaceSet
	}
	tests := []struct{
		name string
		in input
		want NamespaceSet
	}{
		{
			name: "EmptySets",
			in: input{
				left: make(NamespaceSet),
				right: make(NamespaceSet),
			},
			want: make(NamespaceSet),
		},
		{
			name: "EmptyLeft/MultipleRight/NoIntersection",
			in: input{
				left: make(NamespaceSet),
				right: NamespaceSet{
					"ns": {},
					"ns-1": {},
					"ns-2": {},
				},
			},
			want: make(NamespaceSet),
		},
		{
			name: "MultipleLeft/EmptyRight/NoIntersection",
			in: input{
				left: NamespaceSet{
					"ns": {},
					"ns-1": {},
					"ns-2": {},
				},
				right: make(NamespaceSet),
			},
			want: make(NamespaceSet),
		},
		{
			name: "OneLeft/OneRight/Intersection",
			in: input{
				left: NamespaceSet{
					"ns": {},
				},
				right: NamespaceSet{
					"ns": {},
				},
			},
			want: NamespaceSet{
				"ns": {},
			},
		},
		{
			name: "MultipleLeft/MultipleRight/SomeIntersect",
			in: input{
				left: NamespaceSet{
					"ns": {},
					"ns-1": {},
					"ns-2": {},
				},
				right: NamespaceSet{
					"ns": {},
					"ns-1": {},
					"ns-3": {},
				},
			},
			want: NamespaceSet{
				"ns": {},
				"ns-1": {},
			},
		},
		{
			name: "MultipleLeft/MultipleRight/AllIntersect",
			in: input{
				left: NamespaceSet{
					"ns": {},
					"ns-1": {},
					"ns-2": {},
				},
				right: NamespaceSet{
					"ns": {},
					"ns-1": {},
					"ns-2": {},
				},
			},
			want: NamespaceSet{
				"ns": {},
				"ns-1": {},
				"ns-2": {},
			},
		},
		{
			name: "AllLeft/MultipleRight/RightIsIntersection",
			in: input{
				left: NamespaceSet{
					"": {},
				},
				right: NamespaceSet{
					"ns": {},
					"ns-1": {},
					"ns-2": {},
				},
			},
			want: NamespaceSet{
				"ns": {},
				"ns-1": {},
				"ns-2": {},
			},
		},
		{
			name: "MultipleLeft/AllRight/LeftIsIntersection",
			in: input{
				left: NamespaceSet{
					"ns": {},
					"ns-1": {},
					"ns-2": {},
				},
				right: NamespaceSet{
					"": {},
				},
			},
			want: NamespaceSet{
				"ns": {},
				"ns-1": {},
				"ns-2": {},
			},
		},
		{
			name: "AllLeft/EmptyRight/NoIntersection",
			in: input{
				left: NamespaceSet{
					"": {},
				},
				right: make(NamespaceSet),
			},
			want: make(NamespaceSet),
		},
		{
			name: "EmptyLeft/AllRight/NoIntersection",
			in: input{
				left: make(NamespaceSet),
				right: NamespaceSet{
					"": {},
				},
			},
			want: make(NamespaceSet),
		},
		{
			name: "AllLeft/AllRight/Intersection",
			in: input{
				left: NamespaceSet{
					"": {},
				},
				right: NamespaceSet{
					"": {},
				},
			},
			want: NamespaceSet{
				"": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.EqualValues(t, tt.want, tt.in.left.Intersection(tt.in.right))
		})
	}
}

func buildOperatorGroup(namespace, name string, targets []string, gvks []string) *OperatorGroup {
	return NewOperatorGroup(buildAPIOperatorGroup(namespace, name, targets, gvks))
}

func TestGroupIntersection(t *testing.T) {
	type input struct {
		left OperatorGroupSurface
		right []OperatorGroupSurface
	}
	tests := []struct{
		name string
		in input
		want []OperatorGroupSurface	
	}{
		{
			name: "NoTargets/NilGroups/NoIntersection",
			in: input{
				left: buildOperatorGroup("ns", "empty-group", nil, nil),
				right: nil,
			},
			want: []OperatorGroupSurface{},
		},
		{
			name: "MatchingTarget/SingleOtherGroup/Intersection",
			in: input{
				left: buildOperatorGroup("ns", "group-a", []string{"ns-1"}, nil),
				right: []OperatorGroupSurface{
					buildOperatorGroup("ns-2", "group-b", []string{"ns-1"}, nil),
				},
			},
			want: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "group-b", []string{"ns-1"}, nil),
			},
		},
		{
			name: "TargetIsOperatorNamespace/SingleOtherGroup/Intersection",
			in: input{
				left: buildOperatorGroup("ns", "group-a", []string{"ns-1"}, nil),
				right: []OperatorGroupSurface{
					buildOperatorGroup("ns-2", "group-b", []string{"ns"}, nil),
				},
			},
			want: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "group-b", []string{"ns"}, nil),
			},
		},
		{
			name: "MatchingOperatorNamespaces/SingleOtherGroup/Intersection",
			in: input{
				left: buildOperatorGroup("ns", "group-a", []string{"ns-1"}, nil),
				right: []OperatorGroupSurface{
					buildOperatorGroup("ns", "group-b", []string{"ns-2"}, nil),
				},
			},
			want: []OperatorGroupSurface{
				buildOperatorGroup("ns", "group-b", []string{"ns-2"}, nil),
			},
		},
		{
			name: "MatchingTarget/MultipleOtherGroups/Intersection",
			in: input{
				left: buildOperatorGroup("ns", "group-a", []string{"ns-1"}, nil),
				right: []OperatorGroupSurface{
					buildOperatorGroup("ns-2", "group-b", []string{"ns-1"}, nil),
					buildOperatorGroup("ns-3", "group-c", []string{"ns-1"}, nil),
				},
			},
			want: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "group-b", []string{"ns-1"}, nil),
				buildOperatorGroup("ns-3", "group-c", []string{"ns-1"}, nil),
			},
		},
		{
			name: "NonMatchingTargets/MultipleOtherGroups/NoIntersection",
			in: input{
				left: buildOperatorGroup("ns", "group-a", []string{"ns-1", "ns-2", "ns-3"}, nil),
				right: []OperatorGroupSurface{
					buildOperatorGroup("ns-4", "group-b", []string{"ns-6", "ns-7", "ns-8"}, nil),
					buildOperatorGroup("ns-5", "group-c", []string{"ns-6", "ns-7", "ns-8"}, nil),
				},
			},
			want: []OperatorGroupSurface{},
		},
		{
			name: "AllNamespaces/MultipleTargets/Intersection",
			in: input{
				left: buildOperatorGroup("ns", "group-a", []string{""}, nil),
				right: []OperatorGroupSurface{
					buildOperatorGroup("ns-4", "group-b", []string{"ns-6", "ns-7", "ns-8"}, nil),
					buildOperatorGroup("ns-5", "group-c", []string{"ns-9", "ns-10", "ns-11"}, nil),
					buildOperatorGroup("ns-6", "group-d", []string{"ns-11", "ns-12"}, nil),
				},
			},
			want: []OperatorGroupSurface{
				buildOperatorGroup("ns-4", "group-b", []string{"ns-6", "ns-7", "ns-8"}, nil),
				buildOperatorGroup("ns-5", "group-c", []string{"ns-9", "ns-10", "ns-11"}, nil),
				buildOperatorGroup("ns-6", "group-d", []string{"ns-11", "ns-12"}, nil),
			},
		},
		{
			name: "MatchingTargetAllNamespace/MultipleTargets/Intersection",
			in: input{
				left: buildOperatorGroup("ns", "group-a", []string{"ns-1", "ns-2", "ns-3"}, nil),
				right: []OperatorGroupSurface{
					buildOperatorGroup("ns-4", "group-b", []string{""}, nil),
					buildOperatorGroup("ns-5", "group-c", []string{"ns-9", "ns-10", "ns-11"}, nil),
					buildOperatorGroup("ns-6", "group-d", []string{"ns-11", "ns-12"}, nil),
				},
			},
			want: []OperatorGroupSurface{
				buildOperatorGroup("ns-4", "group-b", []string{""}, nil),
			},
		},
		{
			name: "AllNamespace/MultipleTargets/OneAllNamespace/Intersection",
			in: input{
				left: buildOperatorGroup("ns", "group-a", []string{""}, nil),
				right: []OperatorGroupSurface{
					buildOperatorGroup("ns-4", "group-b", []string{""}, nil),
					buildOperatorGroup("ns-5", "group-c", []string{"ns-9", "ns-10", "ns-11"}, nil),
					buildOperatorGroup("ns-6", "group-d", []string{"ns-11", "ns-12"}, nil),
				},
			},
			want: []OperatorGroupSurface{
				buildOperatorGroup("ns-4", "group-b", []string{""}, nil),
				buildOperatorGroup("ns-5", "group-c", []string{"ns-9", "ns-10", "ns-11"}, nil),
				buildOperatorGroup("ns-6", "group-d", []string{"ns-11", "ns-12"}, nil),
			},
		},
		{
			name: "AllNamespace/AllNamespace/Intersection",
			in: input{
				left: buildOperatorGroup("ns", "group-a", []string{""}, nil),
				right: []OperatorGroupSurface{
					buildOperatorGroup("ns-4", "group-b", []string{""}, nil),
				},
			},
			want: []OperatorGroupSurface{
				buildOperatorGroup("ns-4", "group-b", []string{""}, nil),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.EqualValues(t, tt.want, tt.in.left.GroupIntersection(tt.in.right...))
		})
	}

}

func apiIntersectionReconcilerSuite(t *testing.T, reconciler APIIntersectionReconciler) {
	tests := []struct{
		name string
		add APISet
		group OperatorGroupSurface
		otherGroups []OperatorGroupSurface
		want APIReconciliationResult
	}{
		{
			name: "Empty/NoAPIConflict",
			add: make(APISet),
			group: buildOperatorGroup("ns", "g1", []string{"ns"}, nil),
			otherGroups: nil,
			want: NoAPIConflict,
		},
		{
			name: "NoNamespaceIntersection/APIIntersection/NoAPIConflict",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, []string{"Goose.v1alpha1.birds.com"}),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{"ns-3"}, []string{"Goose.v1alpha1.birds.com"}),
			},
			want: NoAPIConflict,
		},
		{
			name: "NamespaceIntersection/NoAPIIntersection/NoAPIConflict",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, []string{"Goose.v1alpha1.birds.com"}),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{"ns-1"}, []string{"Moose.v1alpha1.mammals.com"}),
			},
			want: NoAPIConflict,
		},
		{
			name: "MultipleNamespaceIntersections/NoAPIIntersection/NoAPIConflict",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, []string{"Goose.v1alpha1.birds.com"}),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{"ns-1"}, []string{"Moose.v1alpha1.mammals.com"}),
				buildOperatorGroup("ns-2", "g1", []string{"ns"}, []string{"Egret.v1alpha1.birds.com"}),
			},
			want: NoAPIConflict,
		},
		{
			name: "SomeNamespaceIntersection/NoAPIIntersection/NoAPIConflict",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
				opregistry.APIKey{Group: "mammals.com", Version: "v1alpha1", Kind: "Moose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1", "ns-2", "ns-3"}, []string{"Goose.v1alpha1.birds.com,Moose.v1alpha1.mammals.com"}),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-7", "g1", []string{"ns-4"}, []string{"Moose.v1alpha1.mammals.com"}),
				buildOperatorGroup("ns-8", "g1", []string{"ns-5"}, []string{"Goose.v1alpha1.birds.com"}),
				buildOperatorGroup("ns-9", "g1", []string{""}, []string{"Goat.v1alpha1.mammals.com"}),
			},
			want: NoAPIConflict,
		},
		{
			name: "AllNamespaceIntersection/NoAPIIntersection/NoAPIConflict",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{""}, []string{"Goose.v1alpha1.birds.com"}),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{"ns-1"}, []string{"Moose.v1alpha1.mammals.com"}),
			},
			want: NoAPIConflict,
		},
		{
			name: "AllNamespaceIntersectionOnOther/NoAPIIntersection/NoAPIConflict",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, []string{"Goose.v1alpha1.birds.com"}),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{""}, []string{"Moose.v1alpha1.mammals.com"}),
			},
			want: NoAPIConflict,
		},
		{
			name: "AllNamespaceInstersectionOnOther/NoAPIIntersection/NoAPIConflict",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{""}, []string{"Goose.v1alpha1.birds.com"}),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{""}, []string{"Moose.v1alpha1.mammals.com"}),
			},
			want: NoAPIConflict,
		},
		{
			name: "NamespaceIntersection/NoAPIIntersection/NoAPIConflict",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, []string{"Goose.v1alpha1.birds.com"}),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{"ns-1"}, nil),
			},
			want: NoAPIConflict,
		},
		{
			name: "NamespaceIntersection/APIIntersection/APIConflict",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, nil),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{"ns-1"}, []string{"Goose.v1alpha1.birds.com"}),
			},
			want: APIConflict,
		},
		{
			name: "AllNamespaceIntersection/APIIntersection/APIConflict",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{""}, nil),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{"ns-1"}, []string{"Goose.v1alpha1.birds.com"}),
			},
			want: APIConflict,
		},
		{
			name: "AllNamespaceIntersectionOnOther/APIIntersection/APIConflict",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, nil),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{""}, []string{"Goose.v1alpha1.birds.com"}),
			},
			want: APIConflict,
		},
		{
			name: "AllNamespaceIntersectionOnBoth/APIIntersection/APIConflict",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{""}, nil),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{""}, []string{"Goose.v1alpha1.birds.com"}),
			},
			want: APIConflict,
		},
		{
			name: "NamespaceIntersection/SomeAPIIntersection/APIConflict",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, nil),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{"ns-1"}, []string{"Moose.v1alpha1.birds.com"}),
				buildOperatorGroup("ns-3", "g1", []string{"ns-1"}, []string{"Goose.v1alpha1.birds.com,Egret.v1alpha1.birds.com"}),
			},
			want: APIConflict,
		},
		{
			name: "NamespaceIntersectionOnOperatorNamespace/SomeAPIIntersection/APIConflict",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, nil),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-3", "g1", []string{"ns"}, []string{"Goose.v1alpha1.birds.com,Egret.v1alpha1.birds.com"}),
			},
			want: APIConflict,
		},

		{
			name: "NoNamespaceIntersection/NoAPIIntersection/AddAPIs",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, nil),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{"ns-2"}, []string{"Goose.v1alpha1.birds.com"}),
			},
			want: AddAPIs,
		},
		{
			name: "NamespaceIntersection/NoAPIIntersection/AddAPIs",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, nil),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{"ns-1"}, []string{"Moose.v1alpha1.mammals.com"}),
			},
			want: AddAPIs,
		},
		{
			name: "OperatorNamespaceIntersection/NoAPIIntersection/AddAPIs",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, nil),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{"ns"}, []string{"Moose.v1alpha1.mammals.com"}),
			},
			want: AddAPIs,
		},
		{
			name: "AllNamespaceIntersection/NoAPIIntersection/AddAPIs",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{""}, nil),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{"ns-1"}, []string{"Moose.v1alpha1.mammals.com"}),
				buildOperatorGroup("ns-3", "g1", []string{"ns-1"}, []string{"Goat.v1alpha1.mammals.com,Egret.v1alpha1.birds.com"}),
			},
			want: AddAPIs,
		},
		{
			name: "AllNamespaceIntersectionOnOthers/NoAPIIntersection/AddAPIs",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, nil),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{""}, []string{"Moose.v1alpha1.mammals.com"}),
				buildOperatorGroup("ns-3", "g1", []string{""}, []string{"Goat.v1alpha1.mammals.com,Egret.v1alpha1.birds.com"}),
			},
			want: AddAPIs,
		},
		{
			name: "AllNamespaceIntersectionOnOthers/NoAPIIntersection/AddAPIs/PrexistingAddition",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
				opregistry.APIKey{Group: "mammals.com", Version: "v1alpha1", Kind: "Cow"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, []string{"Cow.v1alpha1.mammals.com"}),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{""}, []string{"Moose.v1alpha1.mammals.com"}),
				buildOperatorGroup("ns-3", "g1", []string{""}, []string{"Goat.v1alpha1.mammals.com,Egret.v1alpha1.birds.com"}),
			},
			want: AddAPIs,
		},
		{
			name: "NamespaceInstersection/APIIntersection/RemoveAPIs",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, []string{"Goose.v1alpha1.birds.com"}),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{"ns-1"}, []string{"Goose.v1alpha1.birds.com"}),
			},
			want: RemoveAPIs,
		},
		{
			name: "AllNamespaceInstersection/APIIntersection/RemoveAPIs",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{""}, []string{"Goose.v1alpha1.birds.com"}),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{"ns-1"}, []string{"Goose.v1alpha1.birds.com"}),
			},
			want: RemoveAPIs,
		},
		{
			name: "AllNamespaceInstersectionOnOther/APIIntersection/RemoveAPIs",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{""}, []string{"Goose.v1alpha1.birds.com"}),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{""}, []string{"Goose.v1alpha1.birds.com"}),
			},
			want: RemoveAPIs,
		},
		{
			name: "MultipleNamespaceIntersections/APIIntersection/RemoveAPIs",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1"}, []string{"Goose.v1alpha1.birds.com"}),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-2", "g1", []string{"ns-1"}, []string{"Goose.v1alpha1.birds.com"}),
				buildOperatorGroup("ns-2", "g1", []string{"ns"}, []string{"Goose.v1alpha1.birds.com"}),
			},
			want: RemoveAPIs,
		},
		{
			name: "SomeNamespaceIntersection/APIIntersection/RemoveAPIs",
			add: APISet{
				opregistry.APIKey{Group: "birds.com", Version: "v1alpha1", Kind: "Goose"}: {},
				opregistry.APIKey{Group: "mammals.com", Version: "v1alpha1", Kind: "Moose"}: {},
			},
			group: buildOperatorGroup("ns", "g1", []string{"ns-1", "ns-2", "ns-3"}, []string{"Goose.v1alpha1.birds.com,Moose.v1alpha1.mammals.com"}),
			otherGroups: []OperatorGroupSurface{
				buildOperatorGroup("ns-7", "g1", []string{"ns-4"}, []string{"Moose.v1alpha1.mammals.com"}),
				buildOperatorGroup("ns-8", "g1", []string{"ns-5", "ns-3"}, []string{"Goose.v1alpha1.birds.com"}),
				buildOperatorGroup("ns-9", "g1", []string{""}, []string{"Goat.v1alpha1.mammals.com"}),
			},
			want: RemoveAPIs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, reconciler.Reconcile(tt.add, tt.group, tt.otherGroups...))
		})
	}
}
func TestReconcileAPIIntersection(t *testing.T) {
	apiIntersectionReconcilerSuite(t, APIIntersectionReconcileFunc(ReconcileAPIIntersection))
}