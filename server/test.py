import pyodbc

# 程序启动时向数据库查询数据总条数
conn = pyodbc.connect(r'DRIVER={SQL Server Native Client 11.0};SERVER=localhost;DATABASE=Douyin;UID=PUGE;PWD=mmit7750')
c = conn.cursor()
startCount = c.execute("select isnull((select top(1) 1 from DouYin.dbo.SIMPLE where DOUYIN_ID = '3790260'), 0) union ALL select isnull((select top(1) 1 from DouYin.dbo.SIMPLE where DOUYIN_ID = '3790220'), 0) union ALL select isnull((select top(1) 1 from DouYin.dbo.SIMPLE where DOUYIN_ID = '3790220'), 0)")
row = startCount.fetchall()
print(row)
conn.commit()
conn.close()