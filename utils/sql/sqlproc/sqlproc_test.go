package sqlproc

import (
	"strings"
	"testing"
)

type validationCase struct {
	name    string
	input   string
	wantErr bool
}

func TestValidateProcedureDefinition(t *testing.T) {
	cases := []validationCase{
		{
			name: "valid_create_procedure_basic",
			input: `
-- TEST: dbo.usp_Test
CREATE PROCEDURE dbo.usp_Test
AS
SELECT 1
`,
		},
		{
			name: "valid_create_proc_short_form",
			input: `
CREATE PROC dbo.usp_Test
AS
SELECT 1
`,
		},
		{
			name: "valid_alter_procedure",
			input: `
ALTER PROCEDURE dbo.usp_Test
AS
SELECT 1
`,
		},
		{
			name: "valid_create_or_alter_proc",
			input: `
CREATE OR ALTER PROC dbo.usp_Test
AS
SELECT 1
`,
		},
		{
			name: "valid_single_part_name",
			input: `
CREATE PROCEDURE usp_Test
AS
SELECT 1
`,
		},
		{
			name: "valid_bracketed_name",
			input: `
CREATE PROCEDURE [dbo].[usp_Test]
AS
SELECT 1
`,
		},
		{
			name: "valid_local_temp_proc",
			input: `
CREATE PROCEDURE #usp_Test
AS
SELECT 1
`,
		},
		{
			name: "valid_global_temp_proc",
			input: `
CREATE PROCEDURE ##usp_Test
AS
SELECT 1
`,
		},
		{
			name: "valid_parameters_defaults_output",
			input: `
CREATE PROCEDURE dbo.usp_Test
    @Id INT = 0,
    @Name NVARCHAR(50) = 'abc',
    @Flag BIT = 1 OUTPUT
AS
SELECT @Id, @Name, @Flag
`,
		},
		{
			name: "valid_as_inside_string_in_header",
			input: `
CREATE PROCEDURE dbo.usp_Test
    @Name NVARCHAR(20) = 'AS'
AS
SELECT @Name
`,
		},
		{
			name: "valid_as_inside_parameter_name",
			input: `
CREATE PROCEDURE dbo.usp_Test
    @TODAS_PARTIDAS BIT = 0
AS
SELECT @TODAS_PARTIDAS
`,
		},
		{
			name: "valid_begin_end_body",
			input: `
CREATE PROCEDURE dbo.usp_Test
AS
BEGIN
    SELECT 1
END
`,
		},
		{
			name: "valid_if_else_without_begin_end",
			input: `
CREATE PROCEDURE dbo.usp_Test
AS
IF 1 = 1
    SELECT 1
ELSE
    SELECT 2
`,
		},
		{
			name: "valid_nested_case",
			input: `
CREATE PROCEDURE dbo.usp_Test
AS
SELECT CASE
    WHEN 1 = 1 THEN CASE WHEN 2 = 2 THEN 100 ELSE 50 END
    ELSE 0
END AS RESULTADO
`,
		},
		{
			name: "valid_try_catch",
			input: `
CREATE PROCEDURE dbo.usp_Test
AS
BEGIN TRY
    SELECT 1
END TRY
BEGIN CATCH
    SELECT ERROR_MESSAGE()
END CATCH
`,
		},
		{
			name: "valid_cursor",
			input: `
CREATE PROCEDURE dbo.usp_Test
AS
DECLARE @Id INT
DECLARE cur CURSOR FOR
    SELECT Id FROM dbo.Clientes

OPEN cur
FETCH NEXT FROM cur INTO @Id

WHILE @@FETCH_STATUS = 0
BEGIN
    SELECT @Id
    FETCH NEXT FROM cur INTO @Id
END

CLOSE cur
DEALLOCATE cur
`,
		},
		{
			name: "valid_dynamic_sql",
			input: `
CREATE PROCEDURE dbo.usp_Test
AS
DECLARE @sql NVARCHAR(MAX) =
    N'SELECT CASE WHEN 1 = 1 THEN ''BEGIN'' ELSE ''END'' END';
EXEC sp_executesql @sql
`,
		},
		{
			name: "valid_cte",
			input: `
CREATE PROCEDURE dbo.usp_Test
AS
;WITH cte AS (
    SELECT 1 AS Id
    UNION ALL
    SELECT 2
)
SELECT * FROM cte
`,
		},
		{
			name: "valid_transaction_not_begin_end_block",
			input: `
CREATE PROCEDURE dbo.usp_Test
AS
BEGIN TRANSACTION
COMMIT TRANSACTION
`,
		},
		{
			name: "valid_comments",
			input: `
/* comentario */
CREATE PROCEDURE dbo.usp_Test
AS
-- comentario
SELECT 1
`,
		},
		{
			name: "valid_realistic_xml_case",
			input: `
CREATE PROCEDURE util_grabar_privilegios_motivopaleta
    @c_emp CHAR(3),
    @c_xml TEXT
AS
BEGIN
    DECLARE
        @hDoc INT,
        @msgerr VARCHAR(3000)

    SELECT @msgerr = ''

    BEGIN TRANSACTION

    EXEC SP_XML_PREPAREDOCUMENT @hDoc OUTPUT, @C_XML

    DELETE FROM privilegios_motivopaleta
    WHERE idempresa = @c_emp

    INSERT INTO privilegios_motivopaleta(idempresa, idmotivopaleta, idusuario)
    SELECT @c_emp, idmotivo, idusuario
    FROM OPENXML (@hDoc, 'VFPData/privilegios_motivopaleta', 2)
    WITH (
        idempresa CHAR(3),
        idmotivo CHAR(3),
        idusuario CHAR(20)
    ) AS mixml

    IF @@ERROR > 0
    BEGIN
        SET @MSGERR = 'No se puede registrar los privilegios de sucursal/mot. produccion'
        GOTO ERROR
    END

    EXEC SP_XML_REMOVEDOCUMENT @hDoc

    SELECT @msgerr AS MENSAJE

    COMMIT TRANSACTION
    RETURN

ERROR:
    BEGIN
        ROLLBACK TRANSACTION
        SELECT @msgerr AS MENSAJE
    END
END
`,
		},
		{
			name:    "invalid_empty",
			input:   ``,
			wantErr: true,
		},
		{
			name: "invalid_whitespace_only",
			input: `

`,
			wantErr: true,
		},
		{
			name: "invalid_missing_create_or_alter_clause",
			input: `
PROCEDURE dbo.usp_Test
AS
SELECT 1
`,
			wantErr: true,
		},
		{
			name: "invalid_missing_procedure_name",
			input: `
CREATE PROCEDURE
AS
SELECT 1
`,
			wantErr: true,
		},
		{
			name: "invalid_missing_as",
			input: `
CREATE PROCEDURE dbo.usp_Test
SELECT 1
`,
			wantErr: true,
		},
		{
			name: "invalid_missing_body",
			input: `
CREATE PROCEDURE dbo.usp_Test
AS
`,
			wantErr: true,
		},
		{
			name: "invalid_unclosed_begin",
			input: `
CREATE PROCEDURE dbo.usp_Test
AS
BEGIN
SELECT 1
`,
			wantErr: true,
		},
		{
			name: "invalid_lonely_end",
			input: `
CREATE PROCEDURE dbo.usp_Test
AS
END
`,
			wantErr: true,
		},
		{
			name: "invalid_unclosed_case",
			input: `
CREATE PROCEDURE dbo.usp_Test
AS
SELECT CASE WHEN 1 = 1 THEN 1
`,
			wantErr: true,
		},
		{
			name: "invalid_unclosed_string",
			input: `
CREATE PROCEDURE dbo.usp_Test
AS
SELECT 'abc
`,
			wantErr: true,
		},
		{
			name: "invalid_unclosed_comment",
			input: `
CREATE PROCEDURE dbo.usp_Test
AS
/* comentario sin cerrar
`,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateProcedureDefinition(tc.input)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected invalid procedure, got nil error\ninput:\n%s", tc.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected valid procedure, got error: %v\ninput:\n%s", err, tc.input)
			}
		})
	}
}

func TestValidateProcedureDefinition_ErrorFormatting(t *testing.T) {
	t.Run("includes_line_column_and_pointer", func(t *testing.T) {
		input := `
CREATE PROCEDURE dbo.usp_Test
SELECT 1
`

		err := ValidateProcedureDefinition(input)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		msg := err.Error()
		if !strings.Contains(msg, "line") {
			t.Fatalf("expected line info in error, got: %s", msg)
		}
		if !strings.Contains(msg, "column") {
			t.Fatalf("expected column info in error, got: %s", msg)
		}
		if !strings.Contains(msg, "^") {
			t.Fatalf("expected pointer in error, got: %s", msg)
		}
	})

	t.Run("empty_input_message_is_clear", func(t *testing.T) {
		err := ValidateProcedureDefinition("")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		msg := strings.ToLower(err.Error())
		if !strings.Contains(msg, "input is empty") {
			t.Fatalf("expected empty input message, got: %s", err.Error())
		}
	})
}

func TestValidateProcedureDefinition_KnownCurrentWeaknesses(t *testing.T) {
	t.Run("unterminated_block_comment_returns_body_missing_or_comment_error", func(t *testing.T) {
		input := `
CREATE PROCEDURE dbo.usp_Test
AS
/* comment never closes
`

		err := ValidateProcedureDefinition(input)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		msg := strings.ToLower(err.Error())
		if !strings.Contains(msg, "expected procedure body after as") &&
			!strings.Contains(msg, "comment") {
			t.Fatalf("expected body-missing or comment-related error, got: %v", err)
		}
	})
}
