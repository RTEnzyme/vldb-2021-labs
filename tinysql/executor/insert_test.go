// Copyright 2018 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package executor_test

import (
	"fmt"
	"strconv"
	"sync"

	. "github.com/pingcap/check"
	"github.com/pingcap/tidb/parser/terror"
	"github.com/pingcap/tidb/table"
	"github.com/pingcap/tidb/util/testkit"
)

func (s *testSuite3) TestInsertWrongValueForField(c *C) {
	tk := testkit.NewTestKit(c, s.store)
	tk.MustExec("use test")
	tk.MustExec(`drop table if exists t1;`)
	tk.MustExec(`create table t1(a bigint);`)
	_, err := tk.Exec(`insert into t1 values("asfasdfsajhlkhlksdaf");`)
	c.Assert(terror.ErrorEqual(err, table.ErrTruncatedWrongValueForField), IsTrue)
}

func (s *testSuite3) TestAllocateContinuousRowID(c *C) {
	tk := testkit.NewTestKit(c, s.store)
	tk.MustExec(`use test`)
	tk.MustExec(`create table t1 (a int,b int, key I_a(a));`)
	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tk := testkit.NewTestKitWithInit(c, s.store)
			for j := 0; j < 10; j++ {
				k := strconv.Itoa(idx*100 + j)
				sql := "insert into t1(a,b) values (" + k + ", 2)"
				for t := 0; t < 20; t++ {
					sql += ",(" + k + ",2)"
				}
				tk.MustExec(sql)
				q := "select _tidb_rowid from t1 where a=" + k
				fmt.Printf("query: %v\n", q)
				rows := tk.MustQuery(q).Rows()
				c.Assert(len(rows), Equals, 21)
				last := 0
				for _, r := range rows {
					c.Assert(len(r), Equals, 1)
					v, err := strconv.Atoi(r[0].(string))
					c.Assert(err, Equals, nil)
					if last > 0 {
						c.Assert(last+1, Equals, v)
					}
					last = v
				}
			}
		}(i)
	}
	wg.Wait()
}

func (s *testSuite3) TestInsertTimeStamp(c *C) {
	tk := testkit.NewTestKit(c, s.store)
	tk.MustExec(`use test`)
	//tk.MustExec(`create table t1 (a int,b int);`)
	tk.MustExec(`create table t1 (a int,b int, c TIMESTAMP);`)
	//fmt.Println(tk.MustQuery("show tables").ows())
	tk.MustExec(`insert into t1(a, b, c) values (1, 2, ` + `"2018-01-01 03:00:01"` + `)`)
	tk.MustExec(`insert into t1(a, b, c) values (1, 2, ` + `"2022-05-01 03:00:01"` + `)`)
	tk.MustExec(`insert into t1(a, b, c) values (3, 3, ` + `"2022-05-01 03:00"` + `)`)
	//tk.MustExec(`insert into t1(a, b) values (1, 2)`)
	fmt.Println("res: ", tk.MustQuery(`select a, b, c from t1`).Rows())
}

func (s *testSuite3) TestInsertDatetime(c *C) {
	tk := testkit.NewTestKit(c, s.store)
	tk.MustExec("use test")
	tk.MustExec(`create table t1 (a int, b int, c datetime);`)
	tk.MustExec(`insert into t1(a, b, c) values (1, 1, "2018-01-01 03:00:00")`)
	tk.MustExec(`insert into t1(a, b, c) values (2, 2, "2018-01-01 03:00:00")`)
	tk.MustExec(`insert into t1(a, b, c) values (3, 3, "2018-01-01 03:00:00")`)
	fmt.Println("res: ", tk.MustQuery(`select a, b, c from t1`))
}

func (s *testSuite3) TestInsertLongText(c *C) {
	tk := testkit.NewTestKit(c, s.store)
	tk.MustExec("use test")
	tk.MustExec("create table t1 (a int, b int, c longtext);")
	tk.MustExec("insert into t1(a, b, c) values (1, 1, '123456')")
	tk.MustExec("insert into t1(a, b, c) values (1, 1, '123')")
	tk.MustExec("insert into t1(a, b, c) values (3, 3, '123')")
	fmt.Println("res: ", tk.MustQuery(`select a, b, c from t1 where c = '123456'`))
}

func (s *testSuite3) TestInvertedIndex(c *C) {
	tk := testkit.NewTestKit(c, s.store)
	tk.MustExec("use test")
	tk.MustExec("create table t1 (a int, b varchar(256), Index b1(b) USING INVERTED)")
	tk.MustExec("insert into t1(a, b) values(1, 'bcd')")
	tk.MustExec("insert into t1(a, b) values(2, 'efg')")
	fmt.Println("res: ", tk.MustQuery(`select a, b from t1 where b in ('abc')`))
}

func (s *testSuite3) TestFuncCutl(c *C) {
	tk := testkit.NewTestKit(c, s.store)
	tk.MustExec("use test")
	tk.MustExec("create table t1 (a int, b varchar(256))")
	tk.MustExec("insert into t1(a, b) values(1, 'bcd')")
	tk.MustExec("insert into t1(a, b) values(2, 'efg')")
	tk.MustExec("insert into t1(a, b) values(2, 'hig')")
	fmt.Println("res: ", tk.MustQuery(`select a, b from t1 where b search_cutl('bcd')`))
}

func (s *testSuite3) TestOrderItemFunc(c *C) {
	tk := testkit.NewTestKit(c, s.store)
	tk.MustExec("use test")
	tk.MustExec("create table t1 (a varchar(256), b varchar(256))")
	tk.MustExec("insert into t1(a, b) values('1236', 'bcd')")
	tk.MustExec("insert into t1(a, b) values('123', 'efg')")
	tk.MustExec("insert into t1(a, b) values('12345', 'hig')")
	fmt.Println("res: ", tk.MustQuery("select a, b from t1 where b cutl('bcd') order by BM25CMP(b, 'f') desc limit 1"))
}
