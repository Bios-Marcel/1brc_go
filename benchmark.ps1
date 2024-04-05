cd baseline
go build -o ../baseline.exe .
cd ../biosmarcel
go build -o ../biosmarcel.exe .
cd ..

$file = $args[0]
hyperfine --warmup 1 "baseline.exe $file" "biosmarcel.exe $file"

