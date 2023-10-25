package domain

/**
 * This interface represents a filter for repositories
 * @param GetFilter gets the filter
 * @param GetData gets the data which will be filtered
 * @param GetOrder gets the order TO-DO: implement this
 * @param GetLimit gets the limit TO-DO: implement this
 * @param GetOffset gets the offset TO-DO: implement this
 * @param GetFields gets the fields TO-DO: implement this
 * @param GetIncludes gets the includes TO-DO: implement this
 * @param GetExcludes gets the excludes TO-DO: implement this
 * @param GetGroupBy gets the group by TO-DO: implement this
 * @param GetHaving gets the having TO-DO: implement this
**/
type RepositoryFilter interface {
	GetFilter() string
	GetData() []interface{}
}
