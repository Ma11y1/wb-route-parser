package utils

func ArrIntToInterface(arr [][]int) [][]interface{} {
	interfaceArray := make([][]interface{}, len(arr))
	for i, row := range arr {
		interfaceArray[i] = make([]interface{}, len(row))
		for j, val := range row {
			interfaceArray[i][j] = val
		}
	}

	return interfaceArray
}
