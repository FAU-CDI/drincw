package sql

import "fmt"

func ExampleBuilder_location() {
	builder := Builder{
		Selectors: make(map[string]Selector),
	}

	builder.AddSelector("name", "column name")
	builder.AddSelector("pohl", "column pohl")

	s, a := builder.Build("import_table_location")

	fmt.Println("<select>")
	fmt.Println(s)
	fmt.Println("</select>")
	if a != "" {
		fmt.Println("<append>")
		fmt.Println(a)
		fmt.Println("</append>")
	}
	//Output: <select>
	// `import_table_location`.`name` as `name`,
	// `import_table_location`.`pohl` as `pohl`
	// </select>
}

func ExampleBuilder_person() {
	builder := Builder{
		Selectors: make(map[string]Selector),
	}

	builder.AddSelector("name", "column name")
	builder.AddSelector("suffix", "column suffix")
	builder.AddSelector("gender", "column gender")
	builder.AddSelector("religion", "column religion")
	builder.AddSelector("stand", "column stand")

	s, a := builder.Build("import_table_person")

	fmt.Println("<select>")
	fmt.Println(s)
	fmt.Println("</select>")
	if a != "" {
		fmt.Println("<append>")
		fmt.Println(a)
		fmt.Println("</append>")
	}
	//Output: <select>
	//`import_table_person`.`gender` as `gender`,
	//`import_table_person`.`name` as `name`,
	//`import_table_person`.`religion` as `religion`,
	//`import_table_person`.`stand` as `stand`,
	//`import_table_person`.`suffix` as `suffix`
	// </select>
}

func ExampleBuilder_institution() {
	builder := Builder{
		Selectors: make(map[string]Selector),
	}

	builder.AddSelector("name", "column name")
	builder.AddSelector("ortsname", "left-join import_table_location locations location_id id name")

	s, a := builder.Build("import_table_instiution")

	fmt.Println("<select>")
	fmt.Println(s)
	fmt.Println("</select>")
	if a != "" {
		fmt.Println("<append>")
		fmt.Println(a)
		fmt.Println("</append>")
	}
	//Output: <select>
	//`import_table_instiution`.`name` as `name`,
	//`locations`.`name` as `ortsname`
	//</select>
	//<append>
	//LEFT JOIN `import_table_location` AS `locations` ON `import_table_instiution`.`location_id` = `locations`.`id`
	//</append>
}
